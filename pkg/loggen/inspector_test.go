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
	"errors"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/query"
	"github.com/naver/lobster/pkg/lobster/util"
)

const endpoint = "lobster-query:8080"

var (
	lobster   = newClient()
	regexpLog = regexp.MustCompile("\"log\":\"(.+?)\"")
)

func requestQuery() (query.ResponseEntries, error) {
	end := time.Now()
	start := end.Add(-time.Minute)

	return lobster.RequestLogEntries(endpoint, query.Request{
		Namespaces:        []string{"user-namespace"},
		Labels:            []model.Labels{{"app": "loggen"}},
		Page:              1,
		Start:             util.Timestamp{Time: start},
		End:               util.Timestamp{Time: end},
		FilterIncludeExpr: "renamedFile",
	})
}

func TestTimestampOrder(t *testing.T) {
	queryResp, err := requestQuery()
	if err != nil {
		t.Error(err)
		return
	}

	prevTime := time.Time{}
	totalCount := 0

	for _, entry := range queryResp.Contents {

		if entry.Timestamp.Before(prevTime) {
			t.Errorf("broken reliability: %v vs %v", prevTime, entry.Timestamp)
			return
		}
		prevTime = entry.Timestamp
		totalCount = totalCount + 1
	}

	if totalCount == 0 {
		t.Error(errors.New("not inspected"))
		return
	}
	t.Logf("verify %d", totalCount)
}

func TestMissingJsonLogs(t *testing.T) {
	queryResp, err := requestQuery()
	if err != nil {
		t.Error(err)
		return
	}

	index := -1
	totalCount := 0

	for _, entry := range queryResp.Contents {
		matches := regexpLog.FindStringSubmatch(entry.Message)
		if len(matches) < 2 {
			t.Errorf("`log` is not found | %s", entry.Message)
			break
		}
		logPart := matches[1]
		found, _ := strconv.Atoi(string(logPart[0]))
		if index < 0 {
			index = found
		}

		if index != found {
			index = (index + 1) % 10
			t.Logf("[%v] %d / %d ", entry.Timestamp, index, found)
			if found != index {
				t.Errorf("broken reliability at %v\nexpected: %v\nfound: %v", entry.Timestamp, index, found)
				break
			}
		}

		totalCount = totalCount + 1
	}
	t.Logf("verify %d", totalCount)
}

func TestMissingTextLogs(t *testing.T) {
	queryResp, err := requestQuery()
	if err != nil {
		t.Error(err)
		return
	}

	index := -1
	totalCount := 0
	for _, entry := range queryResp.Contents {
		found, _ := strconv.Atoi(string(entry.Message[len(entry.Message)-1]))
		if index < 0 {
			index = found
		}

		if index != found {
			index = (index + 1) % 10
			t.Logf("[%v] %d / %d ", entry.Timestamp, index, found)
			if found != index {
				t.Errorf("broken reliability at %v\nexpected: %v\nfound: %v", entry.Timestamp, index, found)
				break
			}
		}

		totalCount = totalCount + 1
	}
	t.Logf("verify %d", totalCount)
}

func TestLatestLog(t *testing.T) {
	now := time.Now()
	queryResp, err := requestQuery()
	if err != nil {
		t.Error(err)
		return
	}
	entries := queryResp.Contents
	ts := entries[len(entries)-1].Timestamp

	t.Log(ts)
	t.Log(now)
	t.Log(now.Sub(ts).Seconds())
}
