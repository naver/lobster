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

package model

import (
	"bytes"
	"fmt"
	"os"
	"strconv"

	"github.com/golang/glog"
)

const CheckPointFileName = "checkpoint"

type CheckPoint struct {
	FileNum int64
	Offset  int64
}

func NewCheckPointFromFile(blockPath string) (*CheckPoint, error) {
	b, err := os.ReadFile(fmt.Sprintf("%s/%s", blockPath, CheckPointFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return NewCheckPoint(-1, 0), nil
		}
		return nil, err
	}
	token := bytes.Split(b, []byte("\n"))
	if len(token[0]) == 0 {
		glog.Errorf("invalid fileNum in (%s): %s", string(b), blockPath)
		return NewCheckPoint(-1, 0), nil
	}
	fileNum, err := strconv.ParseInt(string(token[0]), 0, 64)
	if err != nil {
		return nil, err
	}

	if len(token[1]) == 0 {
		glog.Errorf("invalid offset in (%s) : %s", string(b), blockPath)
		return NewCheckPoint(-1, 0), nil
	}
	offset, err := strconv.ParseInt(string(token[1]), 0, 64)
	if err != nil {
		return nil, err
	}
	return &CheckPoint{FileNum: fileNum, Offset: offset}, nil
}

func NewCheckPoint(fileNum, offset int64) *CheckPoint {
	return &CheckPoint{FileNum: fileNum, Offset: offset}
}

func (c *CheckPoint) SetOffset(offset int64) {
	c.Offset = offset
}

func (c *CheckPoint) Reset(fileNum int64) {
	c.FileNum = fileNum
	c.Offset = 0
}

func (c CheckPoint) ToBytes() []byte {
	return []byte(fmt.Sprintf("%d\n%d", c.FileNum, c.Offset))
}
