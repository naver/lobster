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
	"github.com/naver/lobster/pkg/lobster/query"
)

type keyFunc func(model.Chunk) interface{}
type seekFunc func(map[string]bool, interface{}) bool

type chunkMatcher struct {
	predicates []predicate
}

func NewChunkMatcher(req query.Request) chunkMatcher {
	predicates := []predicate{}

	if matchers := nameMatchers(req); len(matchers) > 0 {
		predicates = append(predicates, and{matchers})
	}
	if matchers := labelMatchers(req); len(matchers) > 0 {
		predicates = append(predicates, or{matchers})
	}

	return chunkMatcher{predicates}
}

func labelMatchers(req query.Request) []matcher {
	matchers := []matcher{}

	for _, label := range req.Labels {
		if len(label) == 0 {
			continue
		}

		matchers = append(matchers, newMatcher(label.Pairs(), seekByKeyValuePairMap, func(c model.Chunk) interface{} { return c.Labels.PairKeyMap() }))
	}

	return matchers
}

func nameMatchers(req query.Request) []matcher {
	matchers := []matcher{}

	if len(req.Clusters) > 0 {
		matchers = append(matchers, newMatcher(req.Clusters, seekByKeyString, func(c model.Chunk) interface{} { return c.Cluster }))
	}
	if len(req.SetNames) > 0 {
		matchers = append(matchers, newMatcher(req.SetNames, seekByKeyString, func(c model.Chunk) interface{} { return c.SetName }))
	}
	if len(req.Pods) > 0 {
		matchers = append(matchers, newMatcher(req.Pods, seekByKeyString, func(c model.Chunk) interface{} { return c.Pod }))
	}
	if len(req.Containers) > 0 {
		matchers = append(matchers, newMatcher(req.Containers, seekByKeyString, func(c model.Chunk) interface{} { return c.Container }))
	}
	if len(req.Sources) > 0 {
		sources := []string{}
		for _, source := range req.Sources {
			sources = append(sources, source.String())
		}
		matchers = append(matchers, newMatcher(sources, seekByKeyString, func(c model.Chunk) interface{} { return c.Source.String() }))
	}

	return matchers
}

func (c chunkMatcher) IsRequestedChunk(chunk model.Chunk) bool {
	for _, predcate := range c.predicates {
		if !predcate.isMatched(chunk) {
			return false
		}
	}

	return true
}

type matcher struct {
	requestedData map[string]bool
	seekFunc      seekFunc
	keyFunc       keyFunc
}

func newMatcher(values []string, seekFunc seekFunc, keyFunc keyFunc) matcher {
	requestedData := map[string]bool{}

	for _, t := range values {
		if len(t) == 0 {
			continue
		}

		requestedData[t] = true
	}

	return matcher{requestedData, seekFunc, keyFunc}
}

func (m matcher) isMatched(chunk model.Chunk) bool {
	return m.seekFunc(m.requestedData, m.keyFunc(chunk))
}

func seekByKeyString(requestedData map[string]bool, key interface{}) bool {
	return requestedData[key.(string)]
}

func seekByKeyValuePairMap(requestedData map[string]bool, keyValuesPairMap interface{}) bool {
	matchedCnt := 0
	converted := keyValuesPairMap.(map[string]bool)

	for keyValuePair := range requestedData {
		if converted[keyValuePair] {
			matchedCnt = matchedCnt + 1
		}
	}

	return len(requestedData) == matchedCnt
}
