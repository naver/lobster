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
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	handleSeconds = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "lobster_handle_seconds",
		Help: "A time spent to handle query",
	}, []string{labelHandler})

	responseBytes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lobster_response_bytes_total",
		Help: "A bytes for response.",
	}, []string{labelHandler})

	responseCodes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lobster_http_response_total",
		Help: "Error code of response",
	}, []string{labelHandler, labelStatusCode})
)

func RegisterMiddlewareMetrics() {
	prometheus.MustRegister(handleSeconds)
	prometheus.MustRegister(responseBytes)
	prometheus.MustRegister(responseCodes)
}

func ObserveHandleSeconds(path string, seconds float64) {
	handleSeconds.WithLabelValues(path).Observe(seconds)
}

func AddResponseBytes(path string, bytesLength int) {
	responseBytes.WithLabelValues(path).Add(float64(bytesLength))
}

func AddResponseStatus(path string, statusCode int) {
	responseCodes.WithLabelValues(path, fmt.Sprint(statusCode)).Inc()
}
