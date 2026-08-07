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

	"github.com/prometheus/client_golang/prometheus"
)

// collectAndCount returns the number of series currently held by a collector.
//
// client_golang's own prometheus/testutil is unusable here: its promlint dependency
// still references expfmt.FmtText, which prometheus/common v0.48.0 removed.
func collectAndCount(c prometheus.Collector) int {
	ch := make(chan prometheus.Metric, 1024)

	go func() {
		c.Collect(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	return count
}

const (
	testNamespace   = "ns-a"
	testPod         = "pod-a"
	testContainer   = "container-a"
	testSourceType  = "stdstream"
	testSourcePath  = ""
	testHandlerPath = "/api/v1/logs/range"
)

func resetStoreMetrics() {
	blockTotal.Reset()
	tailedBytes.Reset()
	tailedLines.Reset()
	overloaded.Reset()
	flushSeconds.Reset()
	handleSeconds.Reset()
	responseBytes.Reset()
	responseCodes.Reset()
}

func addChunkMetrics(namespace, pod, container, sourceType, sourcePath string) {
	SetSizeOfBlocksInChunk(namespace, pod, container, sourceType, sourcePath, 1)
	AddTailedBytes(namespace, pod, container, sourceType, sourcePath, 1)
	AddTailedLines(namespace, pod, container, sourceType, sourcePath, 1)
	AddOverloadedCount(namespace, pod, container, sourceType, sourcePath, "1MB/s")
	ObserveFlushSeconds(namespace, pod, container, sourceType, sourcePath, 0.1)
}

// A collection metric is attributed to labelTargetNamespace and matched by labelLogNamespace,
// and both hold the namespace of the container the logs came from.
func TestChunkLabelValuesFillBothNamespaces(t *testing.T) {
	labels := chunkLabelValues(testNamespace, testPod, testContainer, testSourceType, testSourcePath)

	for _, label := range []string{labelTargetNamespace, labelLogNamespace} {
		if labels[label] != testNamespace {
			t.Errorf("expected %s=%q, got %q", label, testNamespace, labels[label])
		}
	}
}

// Deletion must not depend on labelTargetNamespace, which is free to diverge from the log namespace.
// Building the series with the full label set also pins the label names the vector declares,
// since prometheus panics on a mismatch.
func TestDeleteIgnoresTargetNamespace(t *testing.T) {
	resetStoreMetrics()
	defer resetStoreMetrics()

	tailedLines.With(prometheus.Labels{
		labelTargetNamespace: "sink-ns",
		labelLogNamespace:    testNamespace,
		labelLogPod:          testPod,
		labelLogContainer:    testContainer,
		labelLogSourceType:   testSourceType,
		labelLogSourcePath:   testSourcePath,
	}).Add(1)

	if count := collectAndCount(tailedLines); count != 1 {
		t.Fatalf("lobster_tailed_lines_total: expected 1 series before delete, got %d", count)
	}

	Delete(testNamespace, testPod, testContainer, testSourceType, testSourcePath)

	if count := collectAndCount(tailedLines); count != 0 {
		t.Errorf("lobster_tailed_lines_total: expected 0 series after delete, got %d", count)
	}
}

func TestDeleteRemovesChunkKeyedSeries(t *testing.T) {
	resetStoreMetrics()
	defer resetStoreMetrics()

	addChunkMetrics(testNamespace, testPod, testContainer, testSourceType, testSourcePath)

	before := map[string]int{
		"lobster_blocks":                  collectAndCount(blockTotal),
		"lobster_tailed_bytes_total":      collectAndCount(tailedBytes),
		"lobster_tailed_lines_total":      collectAndCount(tailedLines),
		"lobster_overloaded_target_total": collectAndCount(overloaded),
		"lobster_flush_seconds":           collectAndCount(flushSeconds),
	}
	for name, count := range before {
		if count != 1 {
			t.Fatalf("%s: expected 1 series before delete, got %d", name, count)
		}
	}

	Delete(testNamespace, testPod, testContainer, testSourceType, testSourcePath)

	after := map[string]int{
		"lobster_blocks":                  collectAndCount(blockTotal),
		"lobster_tailed_bytes_total":      collectAndCount(tailedBytes),
		"lobster_tailed_lines_total":      collectAndCount(tailedLines),
		"lobster_overloaded_target_total": collectAndCount(overloaded),
		"lobster_flush_seconds":           collectAndCount(flushSeconds),
	}
	for name, count := range after {
		if count != 0 {
			t.Errorf("%s: expected 0 series after delete, got %d", name, count)
		}
	}
}

func TestDeleteKeepsMiddlewareSeries(t *testing.T) {
	resetStoreMetrics()
	defer resetStoreMetrics()

	ObserveHandleSeconds(testHandlerPath, 0.1)
	AddResponseBytes(testHandlerPath, 100)
	AddResponseStatus(testHandlerPath, 200)

	Delete(testNamespace, testPod, testContainer, testSourceType, testSourcePath)

	// Middleware metrics are keyed by handler path and carry no chunk dimension,
	// so a chunk deletion must never touch them.
	for name, count := range map[string]int{
		"lobster_handle_seconds":       collectAndCount(handleSeconds),
		"lobster_response_bytes_total": collectAndCount(responseBytes),
		"lobster_http_response_total":  collectAndCount(responseCodes),
	} {
		if count != 1 {
			t.Errorf("%s: expected 1 series to survive, got %d", name, count)
		}
	}
}

func TestDeleteOnlyMatchingChunk(t *testing.T) {
	resetStoreMetrics()
	defer resetStoreMetrics()

	// Two overloaded series for the target chunk, distinguished by the limit label.
	AddOverloadedCount(testNamespace, testPod, testContainer, testSourceType, testSourcePath, "1MB/s")
	AddOverloadedCount(testNamespace, testPod, testContainer, testSourceType, testSourcePath, "10MB/s")
	// One series that must survive.
	addChunkMetrics("ns-b", "pod-b", "container-b", testSourceType, testSourcePath)

	if count := collectAndCount(overloaded); count != 3 {
		t.Fatalf("lobster_overloaded_target_total: expected 3 series before delete, got %d", count)
	}

	Delete(testNamespace, testPod, testContainer, testSourceType, testSourcePath)

	if count := collectAndCount(overloaded); count != 1 {
		t.Errorf("lobster_overloaded_target_total: expected 1 surviving series, got %d", count)
	}
	for name, count := range map[string]int{
		"lobster_blocks":             collectAndCount(blockTotal),
		"lobster_tailed_bytes_total": collectAndCount(tailedBytes),
		"lobster_tailed_lines_total": collectAndCount(tailedLines),
		"lobster_flush_seconds":      collectAndCount(flushSeconds),
	} {
		if count != 1 {
			t.Errorf("%s: expected the other chunk's series to survive, got %d", name, count)
		}
	}
}
