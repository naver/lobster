/*
 * Copyright (c) 2024-present NAVER Corp
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package distributor

import (
	"sync"
	"testing"
	"time"

	"github.com/naver/lobster/pkg/lobster/metrics"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/store"
	"github.com/prometheus/client_golang/prometheus"
)

var registerStoreMetricsOnce sync.Once

// hasBlockSeries reports whether a pod holds a lobster_blocks series on the default registry,
// which is what /metrics scrapes.
func hasBlockSeries(t *testing.T, pod string) bool {
	t.Helper()

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}

	for _, family := range families {
		if family.GetName() != "lobster_blocks" {
			continue
		}

		for _, metric := range family.GetMetric() {
			for _, label := range metric.GetLabel() {
				if label.GetName() == "log_pod" && label.GetValue() == pod {
					return true
				}
			}
		}
	}

	return false
}

func staleBlocks(t *testing.T) float64 {
	t.Helper()

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}

	for _, family := range families {
		if family.GetName() != "lobster_stale_blocks" {
			continue
		}
		if len(family.GetMetric()) != 1 {
			t.Fatalf("lobster_stale_blocks: expected exactly 1 series, got %d", len(family.GetMetric()))
		}
		if labels := family.GetMetric()[0].GetLabel(); len(labels) != 0 {
			t.Fatalf("lobster_stale_blocks: expected no labels, got %d", len(labels))
		}

		return family.GetMetric()[0].GetGauge().GetValue()
	}

	t.Fatal("lobster_stale_blocks: series not found")

	return 0
}

func newMetricsTestChunk(pod string, size int64, podDeleted bool) *model.Chunk {
	return &model.Chunk{
		Id:         pod,
		Namespace:  "ns-a",
		Pod:        pod,
		PodUid:     "uid-" + pod,
		Container:  "container-a",
		Source:     model.Source{Type: model.LogTypeStdStream},
		TempBlock:  &model.TempBlock{Size: size},
		UpdatedAt:  time.Now(),
		StartedAt:  time.Now(),
		Size:       size,
		PodDeleted: podDeleted,
	}
}

func newMetricsTestDistributor(t *testing.T, chunks ...*model.Chunk) Distributor {
	t.Helper()
	registerStoreMetricsOnce.Do(metrics.RegisterStoreMetrics)

	s, err := store.NewStore()
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	for _, chunk := range chunks {
		s.StoreChunk(chunk.Source, chunk.PodUid, chunk.Container, chunk)
	}

	return Distributor{store: s}
}

// A chunk whose pod is gone keeps its blocks until retention expires, but the store releases
// its per-chunk series as soon as the pod disappears. Re-setting it here would resurrect a
// series nothing cleans up again, so the size is folded into the unlabeled aggregate instead.
func TestUpdateMetricsFoldsDeletedPodsIntoStaleBlocks(t *testing.T) {
	d := newMetricsTestDistributor(t,
		newMetricsTestChunk("pod-live", 100, false),
		newMetricsTestChunk("pod-gone", 300, true),
		newMetricsTestChunk("pod-gone-too", 700, true),
	)

	d.updateMetrics()

	if !hasBlockSeries(t, "pod-live") {
		t.Error("lobster_blocks: expected the live pod to keep its series")
	}
	for _, pod := range []string{"pod-gone", "pod-gone-too"} {
		if hasBlockSeries(t, pod) {
			t.Errorf("lobster_blocks: expected no series for deleted pod %s", pod)
		}
	}

	if got := staleBlocks(t); got != 1000 {
		t.Errorf("lobster_stale_blocks: expected 1000, got %v", got)
	}
}

// The aggregate is rewritten every round, so once the deleted pods age out it must fall back
// to zero rather than keeping the last total.
func TestUpdateMetricsReportsZeroStaleBlocks(t *testing.T) {
	d := newMetricsTestDistributor(t, newMetricsTestChunk("pod-only-live", 100, false))

	d.updateMetrics()

	if got := staleBlocks(t); got != 0 {
		t.Errorf("lobster_stale_blocks: expected 0, got %v", got)
	}
}
