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

package model

import (
	"testing"
	"time"
)

func TestSeriesDataMeasureWithinRange(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	at := func(min int) time.Time { return base.Add(time.Duration(min) * time.Minute) }

	seriesData := SeriesData{
		{Samples: []Sample{
			{Timestamp: at(0), Lines: 1, Size: 10},
			{Timestamp: at(1), Lines: 2, Size: 20},
			{Timestamp: at(2), Lines: 4, Size: 40},
		}},
		{Samples: []Sample{
			{Timestamp: at(1), Lines: 8, Size: 80},
			{Timestamp: at(3), Lines: 16, Size: 160},
		}},
	}

	tests := []struct {
		name      string
		start     time.Time
		end       time.Time
		wantLines int64
		wantSize  uint64
	}{
		{
			name:      "whole range",
			start:     at(0),
			end:       at(4),
			wantLines: 31,
			wantSize:  310,
		},
		{
			// The end is exclusive so that a sample on a page boundary is not counted twice.
			name:      "end is exclusive",
			start:     at(0),
			end:       at(2),
			wantLines: 11,
			wantSize:  110,
		},
		{
			name:      "start is inclusive and spans multiple series",
			start:     at(1),
			end:       at(2),
			wantLines: 10,
			wantSize:  100,
		},
		{
			name:  "no sample in range",
			start: at(10),
			end:   at(20),
		},
		{
			name:  "empty range",
			start: at(1),
			end:   at(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines, size := seriesData.MeasureWithinRange(tt.start, tt.end)
			if lines != tt.wantLines {
				t.Errorf("lines = %d, want %d", lines, tt.wantLines)
			}
			if size != tt.wantSize {
				t.Errorf("size = %d, want %d", size, tt.wantSize)
			}
		})
	}
}
