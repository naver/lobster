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

const (
	Filtered = 0
	Read     = 1
	SkipRead = 2
	Done     = 3
)

type Result int

type Filterer interface {
	Filter(string, time.Time) (Result, error)
}

func DoFilter(input string, ts time.Time, filterers ...Filterer) (Result, error) {
	for _, filterer := range filterers {
		result, err := filterer.Filter(input, ts)
		if err != nil {
			return result, err
		}
		if result != Read {
			return result, nil
		}
	}
	return Read, nil
}
