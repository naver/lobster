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
	"time"

	"github.com/naver/lobster/pkg/lobster/model"
)

type LimitFunc func(chunk *model.Chunk)

func LimitChunkSize(retentionSize int64) LimitFunc {
	return func(chunk *model.Chunk) {
		remainder := retentionSize
		blocks := chunk.Blocks
		for i := len(blocks) - 1; i >= 0; i-- {
			if remainder > 0 {
				remainder = remainder - blocks[i].Size
			} else {
				chunk.DeletionMarkInBlock = true
				blocks[i].DeletionMark = true
			}
		}
	}
}

func LimitChunkTime(retentionTime time.Duration) LimitFunc {
	return func(chunk *model.Chunk) {
		now := time.Now()

		if chunk.IsOutdated(retentionTime) {
			chunk.DeletionMark = true
			return
		}

		for _, block := range chunk.Blocks {
			if retentionTime < now.Sub(block.EndedAt) {
				chunk.DeletionMarkInBlock = true
				block.DeletionMark = true
			}
		}
	}
}
