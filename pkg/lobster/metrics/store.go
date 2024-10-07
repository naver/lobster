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
	blockTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lobster_blocks",
		Help: "A blocks total.",
	}, []string{labelTargetNamespace, labelLogPod, labelLogContainer, labelLogSourceType, labelLogSourcePath})

	tailedBytes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lobster_tailed_bytes_total",
		Help: "A bytes of lines tailed.",
	}, []string{labelTargetNamespace, labelLogPod, labelLogContainer, labelLogSourceType, labelLogSourcePath})

	tailedLines = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lobster_tailed_lines_total",
		Help: "A number of lines tailed.",
	}, []string{labelTargetNamespace, labelLogPod, labelLogContainer, labelLogSourceType, labelLogSourcePath})

	overloaded = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lobster_overloaded_target_total",
		Help: "A Number of stoping due to overloaded logs.",
	}, []string{labelTargetNamespace, labelLogPod, labelLogContainer, labelLogSourceType, labelLogSourcePath, labelLimit})

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

func SetSizeOfBlocksInChunk(namespace, pod, container, sourceType, sourcePath string, size float64) {
	blockTotal.WithLabelValues(namespace, pod, container, sourceType, sourcePath).Set(size)
}

func AddTailedBytes(namespace, pod, container, sourceType, sourcePath string, bytesLength float64) {
	tailedBytes.WithLabelValues(namespace, pod, container, sourceType, sourcePath).Add(bytesLength)
}

func AddTailedLines(namespace, pod, container, sourceType, sourcePath string, lines float64) {
	tailedLines.WithLabelValues(namespace, pod, container, sourceType, sourcePath).Add(lines)
}

func AddOverloadedCount(namespace, pod, container, sourceType, sourcePath, limit string) {
	overloaded.WithLabelValues(namespace, pod, container, sourceType, sourcePath, limit).Add(1)
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
	blockTotal.Delete(labelChunk)
	tailedBytes.Delete(labelChunk)
	tailedLines.Delete(labelChunk)
	overloaded.DeletePartialMatch(labelChunk)
	flushSeconds.Delete(labelChunk)

	labelNamespace := prometheus.Labels{
		labelLogNamespace: namespace,
	}
	handleSeconds.Delete(labelNamespace)
	responseBytes.Delete(labelNamespace)
	responseCodes.Delete(labelNamespace)
}
