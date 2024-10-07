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

package loggen

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	labelSource = "source"
	labelReason = "reason"
)

var (
	failureCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lobster_loggen_failure_total",
		Help: "A count of failure of inspection.",
	}, []string{labelSource, labelReason})

	verifiedCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lobster_loggen_verified",
		Help: "A count of verified logs.",
	}, []string{labelSource})

	appearedTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lobster_loggen_appeared_time_seconds",
		Help: "The time it takes to reflect the latest logs",
	}, []string{labelSource})
)

func init() {
	prometheus.MustRegister(failureCount)
	prometheus.MustRegister(verifiedCount)
	prometheus.MustRegister(appearedTime)
}

func addFailure(source, reason string) {
	failureCount.With(prometheus.Labels{
		labelSource: source,
		labelReason: reason,
	}).Add(1)
}

func setVerifiedCount(source string, verified float64) {
	verifiedCount.With(prometheus.Labels{labelSource: source}).Set(verified)
}

func setAppearedTime(source string, value float64) {
	appearedTime.With(prometheus.Labels{labelSource: source}).Set(value)
}
