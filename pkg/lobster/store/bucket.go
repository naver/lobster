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
	"fmt"
	"time"
)

const (
	LimitedBySize int = iota
	LimitedByLine
)

type leakyBucket struct {
	limiter  Limiter
	limit    *Limit
	size     int64
	lines    int64
	interval time.Duration
	prevTime time.Time
}

func NewLeakyBucket(limiter Limiter, interval time.Duration) *leakyBucket {
	return &leakyBucket{
		limiter:  limiter,
		limit:    limiter.getDefaultLimit(),
		interval: interval,
		prevTime: time.Time{},
	}
}

func (b *leakyBucket) Init(t time.Time) {
	b.size = 0
	b.lines = 1
	b.limit.release()
	b.limit = b.limiter.getDefaultLimit()
}

func (b *leakyBucket) Pour(size int64) (bool, string) {
	if b.limit.size < b.size {
		b.limit.release()
		b.limit = b.limiter.getLimit(b.size)

		if b.limit.size < b.size {
			return false, description(LimitedBySize, b.limit.size)
		}
	}

	if b.limit.lines < b.lines {
		return false, description(LimitedByLine, b.limit.lines)
	}

	b.size = b.size + size
	b.lines = b.lines + 1

	return true, ""
}

func (b *leakyBucket) Release() {
	b.limit.release()
}

func description(reason int, max int64) string {
	switch reason {
	case LimitedBySize:
		return fmt.Sprintf("Size limit(%d bytes) reached", max)
	case LimitedByLine:
		return fmt.Sprintf("Line count limit(%d) reached", max)
	}
	return "unknown"
}
