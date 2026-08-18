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

package metrics

import (
	"testing"
	"time"

	"github.com/naver/lobster/pkg/lobster/query"
	"github.com/prometheus/client_golang/prometheus"
)

func testExporterLabels(pod string) prometheus.Labels {
	return exporterLabelValues(
		query.Request{
			Namespace: testNamespace,
			Pod:       pod,
			Container: testContainer,
		},
		"sink-ns", "sink-a", "s3", "rule-a")
}

func newTestExpiringMetric() expiringMetric {
	return newExpiringMetricVector(prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lobster_expiring_metric_test_total",
		Help: "A counter used to exercise expiringMetric.",
	}, exporterKeys))
}

func expire(e expiringMetric) {
	for key, hist := range e.historyMap {
		hist.occurred = time.Now().Add(-2 * e.expiration)
		e.historyMap[key] = hist
	}
}

// Go randomizes map iteration order, so a key built by walking the label map must still come
// out identical every time. When it did not, one label set scattered across several history
// entries and stale ones expired series that were still live.
func TestExpiringMetricKeyIsStable(t *testing.T) {
	e := newTestExpiringMetric()
	labels := testExporterLabels(testPod)

	first := e.key(labels)

	for i := 0; i < 500; i++ {
		if got := e.key(labels); got != first {
			t.Fatalf("key is unstable: %q != %q", got, first)
		}
	}
}

// Joining bare values would let two different label sets collide whenever the same values are
// spread across different labels.
func TestExpiringMetricKeyDistinguishesSwappedValues(t *testing.T) {
	e := newTestExpiringMetric()

	a := prometheus.Labels{labelLogPod: "x", labelLogContainer: "y"}
	b := prometheus.Labels{labelLogPod: "y", labelLogContainer: "x"}

	if e.key(a) == e.key(b) {
		t.Errorf("expected distinct keys for swapped values, both are %q", e.key(a))
	}
}

func TestExpiringMetricRefreshKeepsOneEntryPerLabelSet(t *testing.T) {
	e := newTestExpiringMetric()
	labels := testExporterLabels(testPod)

	for i := 0; i < 500; i++ {
		e.Add(labels, 1)
	}

	if len(e.historyMap) != 1 {
		t.Errorf("expected 1 history entry for a single label set, got %d", len(e.historyMap))
	}
	if count := collectAndCount(e.CounterVec); count != 1 {
		t.Errorf("expected 1 series, got %d", count)
	}
}

// A series that keeps being updated must survive, no matter how often it is written.
func TestClearStaleMetricsKeepsRefreshedSeries(t *testing.T) {
	e := newTestExpiringMetric()
	labels := testExporterLabels(testPod)

	for i := 0; i < 100; i++ {
		e.Add(labels, 1)
	}

	e.ClearStaleMetrics()

	if count := collectAndCount(e.CounterVec); count != 1 {
		t.Errorf("expected the refreshed series to survive, got %d series", count)
	}
}

func TestClearStaleMetricsRemovesExpiredSeries(t *testing.T) {
	e := newTestExpiringMetric()

	e.Add(testExporterLabels("pod-stale"), 1)
	expire(e)
	e.Add(testExporterLabels("pod-live"), 1)

	if count := collectAndCount(e.CounterVec); count != 2 {
		t.Fatalf("expected 2 series before clearing, got %d", count)
	}

	e.ClearStaleMetrics()

	if count := collectAndCount(e.CounterVec); count != 1 {
		t.Errorf("expected only the live series to remain, got %d", count)
	}
	if len(e.historyMap) != 1 {
		t.Errorf("expected the expired history entry to be dropped, got %d", len(e.historyMap))
	}
}

// A nil vector is the zero value of the struct; the helpers must stay inert rather than panic.
func TestExpiringMetricIgnoresNilVector(t *testing.T) {
	e := newExpiringMetricVector(nil)

	e.Inc(testExporterLabels(testPod))
	e.Add(testExporterLabels(testPod), 1)

	if len(e.historyMap) != 0 {
		t.Errorf("expected no history to be recorded, got %d entries", len(e.historyMap))
	}
}
