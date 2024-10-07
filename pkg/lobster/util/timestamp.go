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

package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Timestamp struct {
	Time time.Time
}

func (t Timestamp) Add(duration time.Duration) time.Time {
	return t.Time.Add(duration)
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return errors.New("empty timestamp is not allowed")
	}

	var str string

	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	pt, err := ConvertStringToTimestamp(str)
	if err != nil {
		return err
	}

	t.Time = pt.Time
	return nil
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	// this may truncate second precision
	return []byte(fmt.Sprintf("\"%d\"", t.Time.Unix())), nil
}
