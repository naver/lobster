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

package loggen

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/query"
	"github.com/naver/lobster/pkg/lobster/util"
)

type inspectionFunc func(*log.Logger, string, []model.Entry, time.Time, time.Time) bool

var (
	testCases = map[string]inspectionFunc{
		"FailedToGetLogs":       inspectLogAppearTime, // this should be on top to compare an entry's timestamp with the current time
		"InvalidTimestampOrder": inspectTimestampOrder,
		"MissingTextLogs":       inspectMissingTextLogs,
	}
)

func RunInspector(conf Config, req query.Request, stopChan chan struct{}) {
	logger := log.New(os.Stderr, "[inspector]", 0)
	startOffset := time.Duration(0)
	ticker := time.NewTicker(*conf.Interval)
	source := req.FilterIncludeExpr
	lobster := newClient()

	go func() {
		time.Sleep(*conf.WarmUpWait)
		for {
			select {
			case <-ticker.C:
				end := time.Now()
				start := end.Add(-time.Minute).Add(-startOffset)
				req.Start = util.Timestamp{Time: start}
				req.End = util.Timestamp{Time: end}
				queryResp, err := lobster.RequestLogEntries(*conf.LobsterQueryEndpoint, req)
				if err != nil {
					addFailure(source, err.Error())
					logger.Println(err.Error())
					if startOffset < 10*time.Minute {
						startOffset = startOffset + *conf.Interval
					}
					continue
				}

				startOffset = time.Duration(0)

				for reason, fn := range testCases {
					if ok := fn(logger, source, queryResp.Contents, start, end); !ok {
						addFailure(source, reason)
						break
					}
				}
			case <-stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func inspectTimestampOrder(logger *log.Logger, source string, entries []model.Entry, start, end time.Time) bool {
	prevTime := time.Time{}
	totalCount := 0

	for _, e := range entries {
		if e.Timestamp.Before(prevTime) {
			return false
		}
		prevTime = e.Timestamp
		totalCount = totalCount + 1
	}

	return totalCount != 0
}

func inspectMissingTextLogs(logger *log.Logger, source string, entries []model.Entry, start, end time.Time) bool {
	index := -1
	totalCount := 0

	for _, e := range entries {
		if len(e.Message) == 0 {
			if index < 0 {
				// skip
				return true
			}
			continue
		}

		found, err := strconv.Atoi(string(e.Message[len(e.Message)-1]))
		if err != nil {
			// skip
			return true
		}
		if index < 0 {
			index = found
		}

		if index != found {
			index = (index + 1) % 10
			if found != index {
				logger.Printf("broken reliability at %v\nexpected: %v\nfound: %v", e.Timestamp, index, found)
				return false
			}
		}
		totalCount = totalCount + 1
	}

	setVerifiedCount(source, float64(totalCount))

	return true
}

func inspectLogAppearTime(logger *log.Logger, source string, entries []model.Entry, start, end time.Time) bool {
	if len(entries) == 0 {
		return false
	}

	setAppearedTime(source, end.Sub(entries[len(entries)-1].Timestamp).Seconds())

	return true
}
