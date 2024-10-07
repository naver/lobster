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

package tailer

import (
	"flag"
	"time"
)

type config struct {
	ShowTailLog           *bool
	TimeToLive            *time.Duration
	MinStaleTime          *time.Duration
	WaitTimeAfterRotation *time.Duration
}

func setup() config {
	showTailLog := flag.Bool("tailer.showTailLog", false, "Print log of tailing module")
	ttl := flag.Duration("tailer.timeToLive", 10*time.Second, "Log file inspection interval")
	minStaleTime := flag.Duration("tailer.minStaleTime", time.Hour, "Logs older than minStaleTime are considered to be discarded")
	waitTimeAfterRotation := flag.Duration("tailer.waitTimeAfterRotation", 100*time.Millisecond, "Maximum waiting time to collect logs for a rotated file after receiving a rename event.")

	return config{
		ShowTailLog:           showTailLog,
		TimeToLive:            ttl,
		MinStaleTime:          minStaleTime,
		WaitTimeAfterRotation: waitTimeAfterRotation,
	}
}
