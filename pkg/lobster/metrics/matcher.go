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
	matcherKeys = promLabelsKeys(emptyMatcherLabelValues())
	matchedLogs = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lobster_log_metric_matched_logs_total",
		Help: "A number of logs matched.",
	}, matcherKeys)

	matchedLogsError = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lobster_log_metric_matched_logs_error_total",
		Help: "A number of errors during log matches.",
	}, matcherKeys)
)

func RegisterMatcherMetrics() {
	prometheus.MustRegister(matchedLogs)
	prometheus.MustRegister(matchedLogsError)
}

func AddMatchedLogs(req query.Request, sinkNamespace, sinkName, ruleName string) {
	matchedLogs.With(matcherLabelValues(req, sinkNamespace, sinkName, ruleName)).Inc()
}

func AddMatchedLogsError(req query.Request, sinkNamespace, sinkName, ruleName string) {
	matchedLogsError.With(matcherLabelValues(req, sinkNamespace, sinkName, ruleName)).Inc()
}

func DeleteMatchedLogs(namespace, pod, container, sourceType, sourcePath string) {
	partialLabels := prometheus.Labels{
		labelLogNamespace:  namespace,
		labelLogPod:        pod,
		labelLogContainer:  container,
		labelLogSourceType: sourceType,
		labelLogSourcePath: sourcePath,
	}
	matchedLogs.DeletePartialMatch(partialLabels)
	matchedLogsError.DeletePartialMatch(partialLabels)
}

func matcherLabelValues(req query.Request, sinkNamespace, sinkName, ruleName string) prometheus.Labels {
	return prometheus.Labels{
		labelTargetNamespace:  sinkNamespace,
		labelLogNamespace:     req.Namespace,
		labelLogPod:           req.Pod,
		labelLogContainer:     req.Container,
		labelLogSourceType:    req.Source.Type,
		labelLogSourcePath:    req.Source.Path,
		labelSinkContentsName: ruleName,
		labelSinkName:         sinkName,
		labelSinkNamespace:    sinkNamespace,
	}
}

func emptyMatcherLabelValues() prometheus.Labels {
	return matcherLabelValues(query.Request{}, "", "", "")
}
