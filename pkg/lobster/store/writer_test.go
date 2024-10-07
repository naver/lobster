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

package store

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/naver/lobster/pkg/lobster/logline"
)

func TestWriteBufferInsertOutOfOrderReal(t *testing.T) {
	buffer := emptyWriteBuffer()
	istioLogs := []string{
		"2024-01-24T01:01:09.334Z \"GET /xxx\n",
		"2024-01-24T01:01:09.354Z \"POST /xxx\n",
		"2024-01-24T01:01:09.381Z \"GET /xxx\n",
		"2024-01-24T01:01:09.411Z \"GET /xxx\n",
		"2024-01-24T01:01:09.430Z \"POST /xxx\n",
		"2024-01-24T01:01:09.421Z \"GET /xxx\n",

		"2024-01-24T01:01:10.034Z \"GET /xxx\n",
		"2024-01-24T01:01:10.154Z \"POST /xxx\n",
		"2024-01-24T01:01:10.281Z \"GET /xxx\n",
		"2024-01-24T01:01:09.522Z \"GET /xxx\n",
		"2024-01-24T01:01:10.330Z \"POST /xxx\n",
		"2024-01-24T01:01:10.421Z \"GET /xxx\n",
	}

	for _, log := range istioLogs {
		ts, err := logline.ParseTimestampTest(log, logline.LogFormatText)
		if err != nil {
			t.Fatal(err)
		}

		buffer.write(ts, log)
	}

	checkLogOrderInBuffer(buffer, t)
}

func TestWriteBufferInsertNormal(t *testing.T) {
	testCount := 10
	buffer := emptyWriteBuffer()

	for i := 0; i < testCount; i++ {
		ts := time.Now()
		buffer.write(ts, fmt.Sprintf("%s test\n", ts.Format(time.RFC3339Nano)))
	}

	checkLogOrderInBuffer(buffer, t)
}

func TestWriteBufferInsertOutOfOrderFront(t *testing.T) {
	testCount := 10
	expectedLogMsg := "out-of-order start test"
	buffer := emptyWriteBuffer()
	outOfOrderTsFront := time.Now()

	for i := 0; i < testCount; i++ {
		ts := time.Now()
		buffer.write(ts, fmt.Sprintf("%s test\n", ts.Format(time.RFC3339Nano)))
	}

	checkLogOrderInBuffer(buffer, t)

	for i := 0; i < testCount; i++ {
		buffer.write(outOfOrderTsFront, fmt.Sprintf("%s %s\n", outOfOrderTsFront.Format(time.RFC3339Nano), expectedLogMsg))
		checkLogOrderInBuffer(buffer, t)
	}

	logs := strings.Split(buffer.string(), "\n")

	for i := 0; i < testCount; i++ {
		t.Log(logs[i])
		if !strings.Contains(logs[i], expectedLogMsg) {
			t.FailNow()
		}
	}
}
func TestWriteBufferInsertOutOfOrderMiddle(t *testing.T) {
	testCount := 10
	expectedLogMsg := "out-of-order middle test"
	buffer := emptyWriteBuffer()
	outOfOrderTsMiddle := time.Now()

	for i := 0; i < testCount; i++ {
		ts := time.Now()
		buffer.write(ts, fmt.Sprintf("%s test\n", ts.Format(time.RFC3339Nano)))
		if testCount/2 == i {
			outOfOrderTsMiddle = ts
		}
	}

	checkLogOrderInBuffer(buffer, t)

	for i := 0; i < testCount; i++ {
		buffer.write(outOfOrderTsMiddle, fmt.Sprintf("%s %s\n", outOfOrderTsMiddle.Format(time.RFC3339Nano), expectedLogMsg))
		checkLogOrderInBuffer(buffer, t)
	}

	logs := strings.Split(buffer.string(), "\n")

	for i := testCount / 2; i < testCount/2+testCount; i++ {
		t.Log(logs[i])
		if !strings.Contains(logs[i], expectedLogMsg) {
			t.FailNow()
		}
	}
}

func TestWriteBufferInsertOutOfOrderComplex(t *testing.T) {
	testCount := 10
	buffer := emptyWriteBuffer()

	var (
		outOfOrderTsFront   = time.Now()
		outOfOrderTsMiddle1 time.Time
		outOfOrderTsMiddle2 time.Time

		expectedLogMsgNormal  = "test"
		expectedLogMsgFront   = "out-of-order start"
		expectedLogMsgMiddle1 = "out-of-order middle 1"
		expectedLogMsgMiddle2 = "out-of-order middle 2"
	)

	for i := 0; i < testCount; i++ {
		ts := time.Now()
		buffer.write(ts, fmt.Sprintf("%s %s\n", ts.Format(time.RFC3339Nano), expectedLogMsgNormal))
		if testCount/2 == i {
			outOfOrderTsMiddle1 = ts
		}

		if testCount/3 == i {
			outOfOrderTsMiddle2 = ts
		}
	}

	checkLogOrderInBuffer(buffer, t)

	for i := 0; i < testCount; i++ {
		t.Logf("insert out-of-order log to the front %s", outOfOrderTsFront.Format(time.RFC3339Nano))
		buffer.write(outOfOrderTsFront, fmt.Sprintf("%s %s\n", outOfOrderTsFront.Format(time.RFC3339Nano), expectedLogMsgFront))
		checkLogOrderInBuffer(buffer, t)

		t.Logf("insert out-of-order log to the middle 1 %s", outOfOrderTsMiddle1.Format(time.RFC3339Nano))
		buffer.write(outOfOrderTsMiddle1, fmt.Sprintf("%s %s\n", outOfOrderTsMiddle1.Format(time.RFC3339Nano), expectedLogMsgMiddle1))
		checkLogOrderInBuffer(buffer, t)

		t.Logf("insert out-of-order log to the middle 2 %s", outOfOrderTsMiddle2.Format(time.RFC3339Nano))
		buffer.write(outOfOrderTsMiddle2, fmt.Sprintf("%s %s\n", outOfOrderTsMiddle2.Format(time.RFC3339Nano), expectedLogMsgMiddle2))
		checkLogOrderInBuffer(buffer, t)
	}

	counts := make([]int, 4)

	logs := strings.Split(buffer.string(), "\n")

	for i, log := range logs {
		t.Logf("[%d] %s", i, log)
		if strings.Contains(logs[i], expectedLogMsgNormal) {
			counts[0]++
		}
		if strings.Contains(logs[i], expectedLogMsgFront) {
			counts[1]++
		}
		if strings.Contains(logs[i], expectedLogMsgMiddle1) {
			counts[2]++
		}
		if strings.Contains(logs[i], expectedLogMsgMiddle2) {
			counts[3]++
		}
	}

	for i, c := range counts {
		if c != testCount {
			t.Fatalf("count at %d should be %d but %d", i, testCount, c)
		}
	}
}

func checkLogOrderInBuffer(buffer *writeBuffer, t *testing.T) {
	logs := strings.Split(buffer.string(), "\n")

	var prevTs time.Time
	for _, log := range logs {
		if len(log) == 0 {
			continue
		}
		// t.Log(log)
		ts, err := logline.ParseTimestampTest(log, logline.LogFormatText)
		if err != nil {
			t.Fatal(err)
		}
		if prevTs.IsZero() {
			prevTs = ts
			continue
		}
		if ts.Before(prevTs) {
			t.Logf("logTs(%s) should not be showing before prevTs(%s) ", prevTs.Format(time.RFC3339Nano), ts.Format(time.RFC3339Nano))
			t.Fatal(errors.New("inavlid timestamp order"))
		}
		prevTs = ts
	}
}
