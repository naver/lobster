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
	"github.com/prometheus/client_golang/prometheus"
)

var (
	chunkKeys          = promLabelsKeys(emptyChunkLabelValues())
	chunkKeysWithLimit = append(promLabelsKeys(emptyChunkLabelValues()), labelLimit)

	blockTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lobster_blocks",
		Help: "A blocks total.",
	}, chunkKeys)

	staleBlockTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lobster_stale_blocks",
		Help: "A blocks total of chunks whose pod no longer exists.",
	}, []string{})

	tailedBytes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lobster_tailed_bytes_total",
		Help: "A bytes of lines tailed.",
	}, chunkKeys)

	tailedLines = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lobster_tailed_lines_total",
		Help: "A number of lines tailed.",
	}, chunkKeys)

	overloaded = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lobster_overloaded_target_total",
		Help: "A Number of stoping due to overloaded logs.",
	}, chunkKeysWithLimit)

	pushError = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lobster_push_errors_total",
		Help: "An error occurred during pushing",
	}, []string{})

	capOflimit = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lobster_tailed_lines_limit_capacity",
		Help: "Capacity of limits",
	}, []string{labelLimit})

	usageOflimit = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lobster_tailed_lines_limit_usage",
		Help: "Usage of limits",
	}, []string{labelLimit})

	flushSeconds = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "lobster_flush_seconds",
		Help: "A time spent to write file",
	}, []string{labelLogNamespace, labelLogPod, labelLogContainer, labelLogSourceType, labelLogSourcePath})

	rootDiskUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lobster_disk_used",
		Help: "disk usage",
	}, []string{})

	rootDiskLimit = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lobster_disk_limit",
		Help: "disk limit",
	}, []string{})
)

func RegisterStoreMetrics() {
	prometheus.MustRegister(blockTotal)
	prometheus.MustRegister(staleBlockTotal)
	prometheus.MustRegister(tailedBytes)
	prometheus.MustRegister(tailedLines)
	prometheus.MustRegister(overloaded)
	prometheus.MustRegister(pushError)
	prometheus.MustRegister(capOflimit)
	prometheus.MustRegister(usageOflimit)
	prometheus.MustRegister(flushSeconds)
	prometheus.MustRegister(rootDiskUsage)
	prometheus.MustRegister(rootDiskLimit)
}

func chunkLabelValues(namespace, pod, container, sourceType, sourcePath string) prometheus.Labels {
	return prometheus.Labels{
		labelTargetNamespace: namespace,
		labelLogNamespace:    namespace,
		labelLogPod:          pod,
		labelLogContainer:    container,
		labelLogSourceType:   sourceType,
		labelLogSourcePath:   sourcePath,
	}
}

func emptyChunkLabelValues() prometheus.Labels {
	return chunkLabelValues("", "", "", "", "")
}

func SetSizeOfBlocksInChunk(namespace, pod, container, sourceType, sourcePath string, size float64) {
	blockTotal.With(chunkLabelValues(namespace, pod, container, sourceType, sourcePath)).Set(size)
}

func SetSizeOfStaleBlocks(size float64) {
	staleBlockTotal.WithLabelValues().Set(size)
}

func AddTailedBytes(namespace, pod, container, sourceType, sourcePath string, bytesLength float64) {
	tailedBytes.With(chunkLabelValues(namespace, pod, container, sourceType, sourcePath)).Add(bytesLength)
}

func AddTailedLines(namespace, pod, container, sourceType, sourcePath string, lines float64) {
	tailedLines.With(chunkLabelValues(namespace, pod, container, sourceType, sourcePath)).Add(lines)
}

func AddOverloadedCount(namespace, pod, container, sourceType, sourcePath, limit string) {
	labels := chunkLabelValues(namespace, pod, container, sourceType, sourcePath)
	labels[labelLimit] = limit

	overloaded.With(labels).Add(1)
}

func AddPushError() {
	pushError.WithLabelValues().Inc()
}

func SetCapacityOfLimit(cap float64, limit string) {
	capOflimit.WithLabelValues(limit).Set(cap)
}

func SetUsageOfLimit(used float64, limit string) {
	usageOflimit.WithLabelValues(limit).Set(used)
}

func ObserveFlushSeconds(namespace, pod, container, sourceType, sourcePath string, seconds float64) {
	flushSeconds.WithLabelValues(namespace, pod, container, sourceType, sourcePath).Observe(seconds)
}

func SetDiskUsed(size float64) {
	rootDiskUsage.WithLabelValues().Set(size)
}

func SetDiskLimit(size float64) {
	rootDiskLimit.WithLabelValues().Set(size)
}

func Delete(namespace, pod, container, sourceType, sourcePath string) {
	labelChunk := prometheus.Labels{
		labelLogNamespace:  namespace,
		labelLogPod:        pod,
		labelLogContainer:  container,
		labelLogSourceType: sourceType,
		labelLogSourcePath: sourcePath,
	}
	blockTotal.DeletePartialMatch(labelChunk)
	tailedBytes.DeletePartialMatch(labelChunk)
	tailedLines.DeletePartialMatch(labelChunk)
	overloaded.DeletePartialMatch(labelChunk)
	flushSeconds.DeletePartialMatch(labelChunk)
}
