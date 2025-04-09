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

package matcher

import (
	"time"

	"github.com/naver/lobster/pkg/lobster/metrics"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/query/filter"
	"github.com/naver/lobster/pkg/lobster/sink/manager"
	sinkV1 "github.com/naver/lobster/pkg/operator/api/v1"
)

type LogMatcher struct {
	sinkManager manager.SinkManager
}

func NewLogMatcher() LogMatcher {
	return LogMatcher{manager.NewSinkManager(sinkV1.LogMetricRules)}
}

func (m *LogMatcher) Match(key, logLine string, logTs time.Time) {
	orders, ok := m.sinkManager.Load(key)
	if !ok {
		return
	}

	for _, order := range orders {
		result, err := filter.DoFilter(logLine, logTs, order.Request.Filterers...)
		if err != nil {
			metrics.AddMatchedLogsError(order.Request, order.SinkNamespace, order.SinkName, order.RuleName)
		}

		if result != filter.Read {
			continue
		}

		metrics.AddMatchedLogs(order.Request, order.SinkNamespace, order.SinkName, order.RuleName)
	}
}

func (m *LogMatcher) Update(chunks []model.Chunk, start, end time.Time) error {
	return m.sinkManager.Update(chunks, start, end)
}
