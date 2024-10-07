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
	"sync"
)

type Limit struct {
	cap         int
	used        int
	size        int64
	lines       int64
	description string
	sync.RWMutex
}

func newLimit(cap int, maxSize, maxLines int64, description string) *Limit {
	return &Limit{cap, 0, maxSize, maxLines, description, sync.RWMutex{}}
}

func (l *Limit) Stat() (int, int, int64, int64, string) {
	l.RLock()
	defer l.RUnlock()

	return l.cap, l.used, l.size, l.lines, l.description
}

func (l *Limit) use() {
	l.Lock()
	defer l.Unlock()

	l.used = l.used + 1
}

func (l *Limit) useIfAvailable() bool {
	l.Lock()
	defer l.Unlock()

	if l.used < l.cap {
		l.used = l.used + 1
		return true
	}

	return false
}

func (l *Limit) release() {
	l.Lock()
	defer l.Unlock()

	l.used = l.used - 1
}

type Limiter struct {
	limits []*Limit
}

func NewLimiter() Limiter {
	limits := []*Limit{
		newLimit(999, 1000000, 30000, "1MB/s | 30k lines/s"),
		newLimit(30, 20000000, 30000, "20MB/s | 30k lines/s"),
		newLimit(30, 30000000, 30000, "30MB/s | 30k lines/s"),
	}

	if len(limits) == 0 {
		panic("empty pool is not allowed")
	}

	return Limiter{limits}
}

func (l Limiter) GetLimits() []*Limit {
	return l.limits
}

func (l Limiter) getLimit(current int64) *Limit {
	idx := len(l.limits) - 1

	for i := 0; i < len(l.limits); i++ {
		if current <= l.limits[i].size {
			idx = i
			break
		}
	}

	for i := idx; i < len(l.limits); i++ {
		if l.limits[i].useIfAvailable() {
			return l.limits[i]
		}
	}

	for i := idx; i >= 0; i-- {
		if l.limits[i].useIfAvailable() {
			return l.limits[i]
		}
	}

	return l.getDefaultLimit()
}

func (l Limiter) getDefaultLimit() *Limit {
	l.limits[0].use()
	return l.limits[0]
}
