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

package querier

import (
	"github.com/naver/lobster/pkg/lobster/model"
)

type predicate interface {
	isMatched(model.Chunk) bool
}

type and struct {
	matchers []matcher
}

func (a and) isMatched(chunk model.Chunk) bool {
	for _, matcher := range a.matchers {
		if !matcher.isMatched(chunk) {
			return false
		}
	}

	return true
}

type or struct {
	matchers []matcher
}

func (o or) isMatched(chunk model.Chunk) bool {
	for _, matcher := range o.matchers {
		if matcher.isMatched(chunk) {
			return true
		}
	}

	return false
}
