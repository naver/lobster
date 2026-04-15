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

package v1_test

import (
	"reflect"
	"testing"

	sinkV1 "github.com/naver/lobster/pkg/operator/api/v1"
	v1 "github.com/naver/lobster/pkg/operator/server/api/v1"
)

func boolPtr(b bool) *bool { return &b }

func TestMergeRules_LogExportRule(t *testing.T) {
	tests := []struct {
		name     string
		origin   []sinkV1.LogExportRule
		new      []sinkV1.LogExportRule
		expected []sinkV1.LogExportRule
	}{
		{
			name:     "rule present only in origin is appended to result",
			origin:   []sinkV1.LogExportRule{{Name: "keep-me"}},
			new:      []sinkV1.LogExportRule{{Name: "other"}},
			expected: []sinkV1.LogExportRule{{Name: "other"}, {Name: "keep-me"}},
		},
		{
			name:     "rule present only in new is included as-is",
			origin:   []sinkV1.LogExportRule{},
			new:      []sinkV1.LogExportRule{{Name: "new-rule"}},
			expected: []sinkV1.LogExportRule{{Name: "new-rule"}},
		},
		{
			name:   "EnableLogEntryFormat=true in origin is preserved when new omits the field (nil)",
			origin: []sinkV1.LogExportRule{{Name: "foo", EnableLogEntryFormat: boolPtr(true)}},
			new:    []sinkV1.LogExportRule{{Name: "foo"}}, // EnableLogEntryFormat not specified
			expected: []sinkV1.LogExportRule{
				{Name: "foo", EnableLogEntryFormat: boolPtr(true)},
			},
		},
		{
			name:     "EnableLogEntryFormat=false explicitly in new overrides origin true",
			origin:   []sinkV1.LogExportRule{{Name: "foo", EnableLogEntryFormat: boolPtr(true)}},
			new:      []sinkV1.LogExportRule{{Name: "foo", EnableLogEntryFormat: boolPtr(false)}},
			expected: []sinkV1.LogExportRule{{Name: "foo", EnableLogEntryFormat: boolPtr(false)}},
		},
		{
			name:   "BasicBucket.ShouldEncodeFileName=true in origin is preserved when new omits the field",
			origin: []sinkV1.LogExportRule{{Name: "foo", BasicBucket: &sinkV1.BasicBucket{Destination: "dst", ShouldEncodeFileName: boolPtr(true)}}},
			new:    []sinkV1.LogExportRule{{Name: "foo", BasicBucket: &sinkV1.BasicBucket{Destination: "dst"}}},
			expected: []sinkV1.LogExportRule{
				{Name: "foo", BasicBucket: &sinkV1.BasicBucket{Destination: "dst", ShouldEncodeFileName: boolPtr(true)}},
			},
		},
		{
			name:     "BasicBucket.ShouldEncodeFileName=false explicitly in new overrides origin true",
			origin:   []sinkV1.LogExportRule{{Name: "foo", BasicBucket: &sinkV1.BasicBucket{Destination: "dst", ShouldEncodeFileName: boolPtr(true)}}},
			new:      []sinkV1.LogExportRule{{Name: "foo", BasicBucket: &sinkV1.BasicBucket{Destination: "dst", ShouldEncodeFileName: boolPtr(false)}}},
			expected: []sinkV1.LogExportRule{{Name: "foo", BasicBucket: &sinkV1.BasicBucket{Destination: "dst", ShouldEncodeFileName: boolPtr(false)}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := v1.MergeRules(tt.origin, tt.new).([]sinkV1.LogExportRule)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("MergeRules() = %+v, want %+v", got, tt.expected)
			}
		})
	}
}
