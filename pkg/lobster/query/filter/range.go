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

package filter

import (
	"time"
)

type RangeFilter struct {
	start time.Time
	end   time.Time
}

func NewRangeFilter(start, end time.Time) *RangeFilter {
	return &RangeFilter{start, end}
}

func (f *RangeFilter) Filter(input string, ts time.Time) (Result, error) {
	if ts.Before(f.start) || ts.After(f.end) {
		return Filtered, nil
	}

	return Read, nil
}
