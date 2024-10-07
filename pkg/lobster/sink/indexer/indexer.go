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

package indexer

import (
	"github.com/naver/lobster/pkg/lobster/model"
)

type ChunkIndexer map[string][]model.Chunk

func New(chunks []model.Chunk) ChunkIndexer {
	snapshot := ChunkIndexer{}

	for _, chunk := range chunks {
		list, ok := snapshot[chunk.Namespace]
		if !ok {
			snapshot[chunk.Namespace] = []model.Chunk{chunk}
			continue
		}

		snapshot[chunk.Namespace] = append(list, chunk)
	}

	return snapshot
}

func (m ChunkIndexer) GetNamespaces() []string {
	namespaces := []string{}

	for ns := range m {
		namespaces = append(namespaces, ns)
	}

	return namespaces
}

func (m ChunkIndexer) GetChunks(namespace string) []model.Chunk {
	return m[namespace]
}
