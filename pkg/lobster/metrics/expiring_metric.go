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
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type history struct {
	labels   prometheus.Labels
	occurred time.Time
}

type expiringMetric struct {
	CounterVec *prometheus.CounterVec
	historyMap map[string]history
	expiration time.Duration
}

func newExpiringMetricVector(counterVec *prometheus.CounterVec) expiringMetric {
	return expiringMetric{
		CounterVec: counterVec,
		historyMap: map[string]history{},
		expiration: 24 * time.Hour,
	}
}

func (e expiringMetric) Inc(labels prometheus.Labels) {
	if e.CounterVec == nil {
		return
	}
	e.CounterVec.With(labels).Inc()
	e.refresh(labels)
}

func (e expiringMetric) Add(labels prometheus.Labels, value float64) {
	if e.CounterVec == nil {
		return
	}
	e.CounterVec.With(labels).Add(value)
	e.refresh(labels)
}

func (e expiringMetric) refresh(labels prometheus.Labels) {
	key := e.key(labels)
	if hist, ok := e.historyMap[key]; !ok {
		e.historyMap[key] = history{labels, time.Now()}
	} else {
		hist.occurred = time.Now()
		e.historyMap[key] = hist
	}
}

func (e expiringMetric) key(labels prometheus.Labels) string {
	values := []string{}

	for _, value := range labels {
		values = append(values, string(value))
	}

	return strings.Join(values, "")
}

func (e *expiringMetric) ClearStaleMetrics() {
	for key, hist := range e.historyMap {
		if time.Since(hist.occurred) < e.expiration {
			continue
		}

		e.CounterVec.Delete(hist.labels)
		delete(e.historyMap, key)
	}
}
