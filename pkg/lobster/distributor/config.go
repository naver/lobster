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

package distributor

import (
	"flag"
	"time"
)

type config struct {
	StdstreamLogRootPath   *string
	EmptyDirLogRootPath    *string
	FileInspectInterval    *time.Duration
	FileInspectMaxStale    *time.Duration
	TailFileMaxStale       *time.Duration
	MatchLookbackMin       *time.Duration
	MetricsInterval        *time.Duration
	ShouldUpdateLogMatcher *bool
}

func setup() config {
	stdstreamLogRootPath := flag.String("distributor.stdstreamLogRootPath", "/var/log/pods", "Path to find container logs")
	emptyDirLogRootPath := flag.String("distributor.emptyDirLogRootPath", "/var/lib/kubelet/pods", "Path to find pod emptydir logs")
	fileInspectInterval := flag.Duration("distributor.fileInspectInterval", time.Second, "Log file inspection interval")
	fileInspectMaxStale := flag.Duration("distributor.fileInspectMaxStale", 6*24*time.Hour, "Decide how old files to look up; This must be less than store.retentionTime")
	tailFileMaxStale := flag.Duration("distributor.tailFileMaxStale", 5*time.Second, "Decide how old files to look up to tailing")
	matchLookbackMin := flag.Duration("distributor.matchLookbackMin", 10*time.Second, "Determine how old the logs will be in metrics")
	metricsInterval := flag.Duration("distributor.metricsInterval", 5*time.Second, "metrics production interval")
	shouldUpdateLogMatcher := flag.Bool("distributor.shouldUpdateLogMatcher", false, "When using the log sink function, set it to true for periodic log sink rule update")

	return config{
		StdstreamLogRootPath:   stdstreamLogRootPath,
		EmptyDirLogRootPath:    emptyDirLogRootPath,
		FileInspectInterval:    fileInspectInterval,
		FileInspectMaxStale:    fileInspectMaxStale,
		TailFileMaxStale:       tailFileMaxStale,
		MatchLookbackMin:       matchLookbackMin,
		MetricsInterval:        metricsInterval,
		ShouldUpdateLogMatcher: shouldUpdateLogMatcher,
	}
}
