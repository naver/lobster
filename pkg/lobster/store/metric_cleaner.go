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

package store

import (
	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/metrics"
	"github.com/naver/lobster/pkg/lobster/model"
)

type metricKey struct {
	namespace  string
	pod        string
	container  string
	sourceType string
	sourcePath string
}

func chunkMetricKey(chunk *model.Chunk) metricKey {
	return metricKey{
		namespace:  chunk.Namespace,
		pod:        chunk.Pod,
		container:  chunk.Container,
		sourceType: chunk.Source.Type,
		sourcePath: chunk.Source.Path,
	}
}

func (s *Store) cleanMetrics() {
	live := map[metricKey]bool{}
	deleted := []*model.Chunk{}

	s.chunkCache.Range(func(_, value any) bool {
		chunk := value.(*model.Chunk)

		if chunk.PodDeleted {
			deleted = append(deleted, chunk)
		} else {
			live[chunkMetricKey(chunk)] = true
			chunk.MetricsCleared = false
		}

		return true
	})

	for _, chunk := range deleted {
		if chunk.MetricsCleared || live[chunkMetricKey(chunk)] {
			continue
		}

		metrics.Delete(chunk.Namespace, chunk.Pod, chunk.Container, chunk.Source.Type, chunk.Source.Path)
		metrics.DeleteMatchedLogs(chunk.Namespace, chunk.Pod, chunk.Container, chunk.Source.Type, chunk.Source.Path)
		chunk.MetricsCleared = true
		glog.V(3).Infof("clear metrics of deleted pod : %s\n", chunk.Id)
	}
}
