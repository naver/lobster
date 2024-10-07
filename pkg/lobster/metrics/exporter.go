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
	"github.com/naver/lobster/pkg/lobster/query"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	exporterKeys = promLabelsKeys(emptyExporterLabelValues())
	sinkFailure  = newExpiringMetricVector(prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lobster_log_sink_failure_total",
		Help: "Sink failure",
	}, exporterKeys))

	sinkLogBytes = newExpiringMetricVector(prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lobster_log_sink_bytes_total",
		Help: "Amount of exported logs",
	}, exporterKeys))

	exporterHandleSeconds = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "lobster_log_exporter_handle_seconds",
		Help: "A time spent to handle log metric",
	}, []string{})
)

func RegisterExporterMetrics() {
	prometheus.MustRegister(sinkFailure.CounterVec)
	prometheus.MustRegister(sinkLogBytes.CounterVec)
	prometheus.MustRegister(exporterHandleSeconds)
}

func AddSinkFailure(req query.Request, sinkNamespace, sinkName, sinkType, contentsName string) {
	sinkFailure.Inc(exporterLabelValues(req, sinkNamespace, sinkName, sinkType, contentsName))
}

func AddSinkLogBytes(req query.Request, sinkNamespace, sinkName, sinkType, contentsName string, bytes float64) {
	sinkLogBytes.Add(exporterLabelValues(req, sinkNamespace, sinkName, sinkType, contentsName), bytes)
}

func ObserveExporterHandleSeconds(seconds float64) {
	exporterHandleSeconds.WithLabelValues().Observe(seconds)
}

func ClearSinkMetrics() {
	sinkLogBytes.ClearStaleMetrics()
	sinkFailure.ClearStaleMetrics()
}

func exporterLabelValues(req query.Request, sinkNamespace, sinkName, sinkType, contentsName string) prometheus.Labels {
	return prometheus.Labels{
		labelTargetNamespace:  sinkNamespace,
		labelLogNamespace:     req.Namespace,
		labelLogPod:           req.Pod,
		labelLogContainer:     req.Container,
		labelLogSourceType:    req.Source.Type,
		labelLogSourcePath:    req.Source.Path,
		labelSinkContentsName: contentsName,
		labelSinkName:         sinkName,
		labelSinkType:         sinkType,
		labelSinkNamespace:    sinkNamespace,
	}
}

func emptyExporterLabelValues() prometheus.Labels {
	return exporterLabelValues(query.Request{}, "", "", "", "")
}
