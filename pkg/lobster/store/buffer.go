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

import "time"

var maxAge = time.Second

type history struct {
	ts     time.Time
	length int
}

type writeBuffer struct {
	histories  []history
	data       []byte
	start      time.Time
	end        time.Time
	lines      int64
	fileOffset int64
	lastOffset int64
}

func emptyWriteBuffer() *writeBuffer {
	return &writeBuffer{[]history{}, []byte{}, time.Time{}, time.Time{}, 0, 0, 0}
}

func (w *writeBuffer) write(ts time.Time, input string) {
	historyIdx, dataIdx, shouldReorderByTimestamp := w.inspect(ts)
	newHistory := history{ts, len(input)}
	tailLength := len(w.data) - dataIdx

	w.data = append(w.data, []byte(input)...)
	w.histories = append(w.histories, newHistory)

	if shouldReorderByTimestamp {
		copy(w.data[len(w.data)-tailLength:], w.data[dataIdx:dataIdx+tailLength])
		copy(w.data[dataIdx:], []byte(input))
		copy(w.histories[historyIdx+1:], w.histories[historyIdx:])
		w.histories[historyIdx] = newHistory
	}

	w.lines = w.lines + 1
	w.fileOffset = w.fileOffset + int64(len(input))

	w.start = w.histories[0].ts
	w.end = w.histories[len(w.histories)-1].ts
}

func (w writeBuffer) inspect(ts time.Time) (int, int, bool) {
	var (
		minTs      = ts.Add(-maxAge)
		historyIdx = len(w.histories) - 1
		dataIdx    = len(w.data)
	)

	if len(w.histories) == 0 || w.histories[len(w.histories)-1].ts.Before(ts) {
		return 0, 0, false
	}

	for ; historyIdx >= 0; historyIdx-- {
		if ts.After(w.histories[historyIdx].ts) || w.histories[historyIdx].ts.Before(minTs) {
			break
		}
		dataIdx = dataIdx - w.histories[historyIdx].length
	}

	return historyIdx + 1, dataIdx, true
}

func (w writeBuffer) isValid() bool {
	return w.start.After(w.end)
}

func (w *writeBuffer) resetFileOffset() {
	w.fileOffset = 0
}

func (w *writeBuffer) reset() {
	w.histories = w.histories[:0]
	w.data = w.data[:0]
	w.start = time.Time{}
	w.end = time.Time{}
	w.lines = 0
}

func (w writeBuffer) size() int {
	return len(w.data)
}

func (w writeBuffer) string() string {
	return string(w.data)
}

func (w writeBuffer) bytes() []byte {
	return w.data
}
