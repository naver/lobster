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

package helper

import (
	"github.com/naver/lobster/pkg/lobster/model"
	v1 "k8s.io/api/core/v1"
)

func FilterChunksByExistingPods(chunks []model.Chunk, podMap map[string]v1.Pod) []model.Chunk {
	var filtered []model.Chunk

	for _, chunk := range chunks {
		if _, ok := podMap[chunk.PodUid]; !ok {
			continue
		}

		filtered = append(filtered, chunk)
	}

	return filtered
}
