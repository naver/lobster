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
	"time"
)

const liveFactor = 2

type Receipt struct {
	ExportBytes    int
	ExportTime     time.Time
	ExportInterval time.Duration
	LogTime        time.Time
}

func (r *Receipt) Update(exportBytes int, exportTime time.Time, interval time.Duration, logTime time.Time) {
	r.ExportBytes = exportBytes
	r.ExportTime = exportTime
	r.ExportInterval = interval
	r.LogTime = logTime
}

func (r Receipt) IsStale(t time.Time) bool {
	return t.Sub(r.ExportTime).Seconds() > liveFactor*r.ExportInterval.Seconds()
}
