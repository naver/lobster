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

package counter

import (
	"testing"
	"time"
)

func TestSafeLookback(t *testing.T) {
	tests := []struct {
		name        string
		interval    time.Duration
		maxLookback time.Duration
		expected    time.Duration
	}{
		{
			name:        "liveFactor * interval is smaller than maxLookback",
			interval:    10 * time.Minute,
			maxLookback: time.Hour,
			expected:    30 * time.Minute, // liveFactor(3) * 10min
		},
		{
			name:        "maxLookback is smaller than liveFactor * interval",
			interval:    30 * time.Minute,
			maxLookback: time.Hour,
			expected:    time.Hour, // maxLookback
		},
		{
			name:        "equal values",
			interval:    20 * time.Minute,
			maxLookback: time.Hour,
			expected:    time.Hour, // both are equal
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeLookback(tt.interval, tt.maxLookback)
			if result != tt.expected {
				t.Errorf("SafeLookback() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSafeLookbackCoversDelayedLogs(t *testing.T) {
	interval := 10 * time.Minute
	maxLookback := time.Hour
	safeLookback := SafeLookback(interval, maxLookback)

	current := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	// Simulate initial receipt creation with SafeLookback
	// ExportTime = current - interval (ensures isRightTimeToExport passes)
	// LogTime = current - safeLookback (determines log query range)
	initialReceiptExportTime := current.Add(-interval)
	initialReceiptLogTime := current.Add(-safeLookback)

	// Verify isRightTimeToExport() would pass for new receipt
	// Logic: interval.Seconds() <= current.Sub(exportTime).Seconds()
	timeSinceExport := current.Sub(initialReceiptExportTime)
	if timeSinceExport < interval {
		t.Errorf("isRightTimeToExport() would fail: timeSinceExport=%v < interval=%v", timeSinceExport, interval)
	}

	// Delayed logs that arrived now but have old timestamps
	delayedLogs := []struct {
		name         string
		logTimestamp time.Time
		shouldExport bool
	}{
		{
			name:         "log from 5 minutes ago",
			logTimestamp: current.Add(-5 * time.Minute),
			shouldExport: true,
		},
		{
			name:         "log from 15 minutes ago (outside interval, inside safeLookback)",
			logTimestamp: current.Add(-15 * time.Minute),
			shouldExport: true,
		},
		{
			name:         "log from 25 minutes ago (outside interval, inside safeLookback)",
			logTimestamp: current.Add(-25 * time.Minute),
			shouldExport: true,
		},
		{
			name:         "log from exactly safeLookback ago",
			logTimestamp: current.Add(-safeLookback),
			shouldExport: true,
		},
		{
			name:         "log from beyond safeLookback",
			logTimestamp: current.Add(-safeLookback - time.Second),
			shouldExport: false,
		},
	}

	for _, log := range delayedLogs {
		t.Run(log.name, func(t *testing.T) {
			// Check if log would be included in export range
			inRange := !log.logTimestamp.Before(initialReceiptLogTime) && !log.logTimestamp.After(current)

			if inRange != log.shouldExport {
				t.Errorf("log export = %v, expected %v (logTimestamp: %v, initialLogTime: %v, safeLookback: %v)",
					inRange, log.shouldExport, log.logTimestamp, initialReceiptLogTime, safeLookback)
			}
		})
	}
}
