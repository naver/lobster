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
	"regexp"
	"time"
)

type NegativeRegexpFilterer struct {
	compiled *regexp.Regexp
}

func NewNegativeRegexpFilterer(expr string) (*NegativeRegexpFilterer, error) {
	if v, ok := compiledRegexpCache.Get(expr); ok {
		return &NegativeRegexpFilterer{v.(*regexp.Regexp)}, nil
	}

	compiled, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}

	compiledRegexpCache.Add(expr, compiled)

	return &NegativeRegexpFilterer{compiled}, nil
}

func (f *NegativeRegexpFilterer) Filter(input string, _ time.Time) (Result, error) {
	if !f.compiled.MatchString(input) {
		return Read, nil
	}

	return Filtered, nil
}
