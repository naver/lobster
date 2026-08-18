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

package store

import (
	"sync"
	"testing"

	"github.com/naver/lobster/pkg/lobster/metrics"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/query"
	"github.com/prometheus/client_golang/prometheus"
)

// Chunk.MetricsCleared flips exactly when metrics.Delete is called, so these tests assert on the
// flag instead of reaching into the metrics package.
// The deletion itself is covered by pkg/lobster/metrics/store_test.go.

var registerMatcherMetricsOnce sync.Once

// countSeries reads the default registry, which is what /metrics scrapes, and counts the
// series of one metric family belonging to a single pod.
func countSeries(t *testing.T, familyName, pod string) int {
	t.Helper()

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}

	count := 0

	for _, family := range families {
		if family.GetName() != familyName {
			continue
		}

		for _, metric := range family.GetMetric() {
			for _, label := range metric.GetLabel() {
				if label.GetName() == "log_pod" && label.GetValue() == pod {
					count++
				}
			}
		}
	}

	return count
}

func newTestStore() *Store {
	return &Store{chunkCache: sync.Map{}}
}

func newTestChunk(pod, podUid string, podDeleted bool) *model.Chunk {
	return &model.Chunk{
		Id:         pod + "_" + podUid,
		Namespace:  "ns-a",
		Pod:        pod,
		PodUid:     podUid,
		Container:  "container-a",
		Source:     model.Source{Type: model.LogTypeStdStream},
		TempBlock:  &model.TempBlock{},
		PodDeleted: podDeleted,
	}
}

func (s *Store) storeTestChunk(chunk *model.Chunk) {
	s.StoreChunk(chunk.Source, chunk.PodUid, chunk.Container, chunk)
}

func TestCleanMetricsClearsDeletedPod(t *testing.T) {
	s := newTestStore()
	chunk := newTestChunk("pod-a", "uid-a", true)
	s.storeTestChunk(chunk)

	s.cleanMetrics()

	if !chunk.MetricsCleared {
		t.Error("expected metrics of the deleted pod to be cleared")
	}
}

func TestCleanMetricsKeepsExistingPod(t *testing.T) {
	s := newTestStore()
	chunk := newTestChunk("pod-a", "uid-a", false)
	s.storeTestChunk(chunk)

	s.cleanMetrics()

	if chunk.MetricsCleared {
		t.Error("expected metrics of the existing pod to be kept")
	}
}

func TestCleanMetricsIsIdempotent(t *testing.T) {
	s := newTestStore()
	chunk := newTestChunk("pod-a", "uid-a", true)
	s.storeTestChunk(chunk)

	s.cleanMetrics()
	s.cleanMetrics()

	if !chunk.MetricsCleared {
		t.Error("expected the cleared state to hold across repeated sweeps")
	}
}

func TestCleanMetricsRearmsWhenPodReappears(t *testing.T) {
	s := newTestStore()
	chunk := newTestChunk("pod-a", "uid-a", true)
	s.storeTestChunk(chunk)

	s.cleanMetrics()
	if !chunk.MetricsCleared {
		t.Fatal("expected metrics of the deleted pod to be cleared")
	}

	// A pod list that briefly missed the pod must not leave the chunk stuck in the cleared state.
	chunk.PodDeleted = false
	s.cleanMetrics()

	if chunk.MetricsCleared {
		t.Error("expected the chunk to be rearmed once the pod reappeared")
	}
}

func TestCleanMetricsSparesAliasedLivePod(t *testing.T) {
	s := newTestStore()
	// A recreated pod keeps its name but takes a new uid, so both chunks share one metric label set.
	old := newTestChunk("pod-a", "uid-old", true)
	recreated := newTestChunk("pod-a", "uid-new", false)
	s.storeTestChunk(old)
	s.storeTestChunk(recreated)

	s.cleanMetrics()

	if old.MetricsCleared {
		t.Error("expected the chunk of the old pod to spare the metrics of its living namesake")
	}
	if recreated.MetricsCleared {
		t.Error("expected metrics of the recreated pod to be kept")
	}
}

func TestCleanMetricsClearsDeletedPodOfDistinctName(t *testing.T) {
	s := newTestStore()
	gone := newTestChunk("pod-a", "uid-a", true)
	alive := newTestChunk("pod-b", "uid-b", false)
	s.storeTestChunk(gone)
	s.storeTestChunk(alive)

	s.cleanMetrics()

	if !gone.MetricsCleared {
		t.Error("expected the deleted pod to be cleared when no namesake is alive")
	}
	if alive.MetricsCleared {
		t.Error("expected metrics of the existing pod to be kept")
	}
}

// cleanMetrics releases every series keyed by the chunk, not just the ones behind
// metrics.Delete. Matched-log series carry the same pod label and used to be released only by
// retention, so they outlived the pod by up to store.retentionTime.
func TestCleanMetricsClearsMatchedLogsOfDeletedPod(t *testing.T) {
	registerMatcherMetricsOnce.Do(metrics.RegisterMatcherMetrics)

	s := newTestStore()
	chunk := newTestChunk("pod-matched", "uid-matched", true)
	s.storeTestChunk(chunk)

	req := query.Request{
		Namespace: chunk.Namespace,
		Pod:       chunk.Pod,
		Container: chunk.Container,
		Source:    chunk.Source,
	}
	metrics.AddMatchedLogs(req, "sink-ns", "sink-a", "rule-a")
	metrics.AddMatchedLogsError(req, "sink-ns", "sink-a", "rule-a")

	if count := countSeries(t, "lobster_log_metric_matched_logs_total", chunk.Pod); count != 1 {
		t.Fatalf("expected 1 matched-logs series before cleaning, got %d", count)
	}

	s.cleanMetrics()

	for _, name := range []string{
		"lobster_log_metric_matched_logs_total",
		"lobster_log_metric_matched_logs_error_total",
	} {
		if count := countSeries(t, name, chunk.Pod); count != 0 {
			t.Errorf("%s: expected the deleted pod's series to be released, got %d", name, count)
		}
	}
}

func TestCleanMetricsKeepsMatchedLogsOfExistingPod(t *testing.T) {
	registerMatcherMetricsOnce.Do(metrics.RegisterMatcherMetrics)

	s := newTestStore()
	chunk := newTestChunk("pod-matched-live", "uid-matched-live", false)
	s.storeTestChunk(chunk)

	metrics.AddMatchedLogs(query.Request{
		Namespace: chunk.Namespace,
		Pod:       chunk.Pod,
		Container: chunk.Container,
		Source:    chunk.Source,
	}, "sink-ns", "sink-a", "rule-a")

	s.cleanMetrics()

	if count := countSeries(t, "lobster_log_metric_matched_logs_total", chunk.Pod); count != 1 {
		t.Errorf("expected the live pod's series to survive, got %d", count)
	}
}
