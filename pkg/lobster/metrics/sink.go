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
	sinkHandleSeconds = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "lobster_log_sink_handle_seconds",
		Help: "A time spent to handle log metric",
	}, []string{})

	receivingPreorders = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lobster_log_sink_pre_orders",
		Help: "A number of receiving preorders.",
	}, []string{})

	sinkRequestFailure = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lobster_log_sink_request_orders_failure_total",
		Help: "A number of faiure of receiving sinks.",
	}, []string{})

	orders = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lobster_log_sink_orders",
		Help: "A number of orders.",
	}, []string{})

	invalidRequest = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lobster_log_sink_invalid_request_total",
		Help: "A number of invalid request.",
	}, []string{labelSinkNamespace, labelSinkName, labelSinkContentsName})
)

func RegisterSinkMetrics() {
	prometheus.MustRegister(sinkHandleSeconds)
	prometheus.MustRegister(receivingPreorders)
	prometheus.MustRegister(sinkRequestFailure)
	prometheus.MustRegister(orders)
	prometheus.MustRegister(invalidRequest)
}

func ObserveSinkHandleSeconds(seconds float64) {
	sinkHandleSeconds.WithLabelValues().Observe(seconds)
}

func SetOrders(num int) {
	orders.WithLabelValues().Set(float64(num))
}

func SetReceivingPreorders(count float64) {
	receivingPreorders.WithLabelValues().Set(count)
}

func AddSinkRequestFailureCount() {
	sinkRequestFailure.WithLabelValues().Inc()
}

func AddInvalidRequestCount(namespace, sink, name string) {
	invalidRequest.WithLabelValues(namespace, sink, name).Inc()
}
