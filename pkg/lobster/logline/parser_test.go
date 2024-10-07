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

package logline

import (
	"testing"
	"time"
)

func TestLogTextLine(t *testing.T) {
	testData := map[string]string{
		`2022-11-01T09:14:12.652952473-09:00 stdout F ...`: `2022-11-02T03:14:12.652952473+09:00`, // k8s stdstream
		`2022-11-01T09:14:12.652952473+09:00 stdout F ...`: `2022-11-01T09:14:12.652952473+09:00`, // k8s stdstream
		`2022-11-01T09:14:12.652952473+07:30 stdout F ...`: `2022-11-01T11:44:12.652952473+09:00`, // k8s stdstream
		`2022-11-01T09:14:12.652952473-07:30 stdout F ...`: `2022-11-02T01:44:12.652952473+09:00`, // k8s stdstream
		`2022-11-01T09:14:12.65471569+09:00 stdout F ...`:  `2022-11-01T09:14:12.65471569+09:00`,  // k8s stdstream
		`2022-11-01T09:14:12.841438+09:00 stdout F ...`:    `2022-11-01T09:14:12.841438+09:00`,    // k8s stdstream
		`2022-11-01T09:14:15.653862+09:00 stdout F ...`:    `2022-11-01T09:14:15.653862+09:00`,    // k8s stdstream
		`2023-12-05T06:52:01.364Z ...`:                     `2023-12-05T15:52:01.364+09:00`,       // istio access log
		`2023-12-05T06:52:01Z ...`:                         `2023-12-05T15:52:01+09:00`,           // istio access log
		`2023-12-12T17:51:43.769+09:00	...`:                `2023-12-12T17:51:43.769+09:00`,       // has tap
		`2023-12-12T17:51:43.769+0900 ...`:                 `2023-12-12T17:51:43.769+09:00`,       // ISO8601
		`2024-05-13T14:57:24+09:00 ...`:                    `2024-05-13T14:57:24+09:00`,           // ratelimit
	}

	for question, expected := range testData {
		ts, err := ParseTimestampTest(question, LogFormatText)
		if err != nil {
			t.Log("question: " + question)
			t.Log("expected: " + expected)
			t.Error(err)
			return
		}
		if ts.Format(time.RFC3339Nano) != expected {
			t.Log("question: " + question)
			t.Log("expected: " + expected)
			t.Log("returned: " + ts.Format(time.RFC3339Nano))
			return
		}
	}
}

func BenchmarkLogTextLine(b *testing.B) {
	testData := "2023-12-05T06:52:01.364Z message"

	for i := 0; i < b.N; i++ {
		_, err := ParseTimestampTest(testData, LogFormatText)
		if err != nil {
			b.Error(err)
			return
		}
	}
}
