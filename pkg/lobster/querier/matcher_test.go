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
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/query"
)

var (
	clusters = []string{"cluster-a", "cluster-b", "cluster-c"}
	chunks   = []*model.Chunk{}
)

func init() {
	for i := 0; i < 20; i++ {
		suffix := fmt.Sprintf("-%d", i)
		chunk, _ := model.NewChunk(model.LogFile{
			Labels:    map[string]string{"label": "label" + suffix},
			Pod:       "pod" + suffix,
			PodUid:    "pod-uid" + suffix,
			Container: "container" + suffix,
			Source:    model.Source{Type: model.LogTypeStdStream, Path: ""},
		}, nil)
		chunks = append(chunks, chunk)
	}

	for i := range chunks {
		chunks[i].Cluster = clusters[i%len(clusters)]
	}
}

func TestEntireChunks(t *testing.T) {
	matcher := NewChunkMatcher(query.Request{})

	matchedChunks := []model.Chunk{}
	for _, chunk := range chunks {
		if matcher.IsRequestedChunk(*chunk) {
			matchedChunks = append(matchedChunks, *chunk)
		}
	}

	if len(matchedChunks) != len(chunks) {
		t.Fail()
	}
}

func TestMatchSingleCluster(t *testing.T) {
	var expectedClusters = []string{"cluster-a"}

	matcher := NewChunkMatcher(query.Request{
		Clusters: expectedClusters,
	})

	matchedChunks := []model.Chunk{}
	for _, chunk := range chunks {
		if matcher.IsRequestedChunk(*chunk) {
			matchedChunks = append(matchedChunks, *chunk)
		}
	}

	for _, matched := range matchedChunks {
		t.Log(matched)
		if !slices.Contains(expectedClusters, matched.Cluster) {
			t.Fail()
		}
	}
}

func TestMatchMultipleClusters(t *testing.T) {
	var expectedClusters = []string{"cluster-a", "cluster-b"}

	matcher := NewChunkMatcher(query.Request{
		Clusters: expectedClusters,
	})

	matchedChunks := []model.Chunk{}
	for _, chunk := range chunks {
		if matcher.IsRequestedChunk(*chunk) {
			matchedChunks = append(matchedChunks, *chunk)
		}
	}

	for _, matched := range matchedChunks {
		t.Log(matched)
		if !slices.Contains(expectedClusters, matched.Cluster) {
			t.Fail()
		}
	}
}

func TestMatchSingleLabel(t *testing.T) {
	var (
		expectedClusters     = []string{"cluster-a"}
		expectedLabels       = []model.Labels{{"label": "label-0"}}
		expectedLabelStrings []string
	)

	for _, label := range expectedLabels {
		expectedLabelStrings = append(expectedLabelStrings, label.String())
	}

	matcher := NewChunkMatcher(query.Request{
		Clusters: expectedClusters,
		Labels:   expectedLabels,
	})

	matchedChunks := []model.Chunk{}
	for _, chunk := range chunks {
		if matcher.IsRequestedChunk(*chunk) {
			matchedChunks = append(matchedChunks, *chunk)
		}
	}

	for _, matched := range matchedChunks {
		t.Log(matched)
		if !slices.Contains(expectedClusters, matched.Cluster) {
			t.Fail()
		}

		hasProperLabel := false
		matchedLabelString := matched.Labels.String()

		for _, labelString := range expectedLabelStrings {
			if strings.Contains(matchedLabelString, labelString) {
				hasProperLabel = true
				break
			}
		}

		if !hasProperLabel {
			t.Fail()
		}
	}
}

func TestMatchMultipleLabels(t *testing.T) {
	var (
		expectedClusters     = []string{"cluster-a"}
		expectedLabels       = []model.Labels{{"label": "label-0"}, {"label": "label-1"}}
		expectedLabelStrings []string
	)

	for _, label := range expectedLabels {
		expectedLabelStrings = append(expectedLabelStrings, label.String())
	}

	matcher := NewChunkMatcher(query.Request{
		Clusters: expectedClusters,
		Labels:   expectedLabels,
	})

	matchedChunks := []model.Chunk{}
	for _, chunk := range chunks {
		if matcher.IsRequestedChunk(*chunk) {
			matchedChunks = append(matchedChunks, *chunk)
		}
	}

	for _, matched := range matchedChunks {
		t.Log(matched)
		if !slices.Contains(expectedClusters, matched.Cluster) {
			t.Fail()
		}

		hasProperLabel := false
		matchedLabelString := matched.Labels.String()

		for _, labelString := range expectedLabelStrings {
			if matchedLabelString == labelString {
				hasProperLabel = true
				break // escape this to check 'or condition'
			}
		}

		if !hasProperLabel {
			t.Fail()
		}
	}
}
