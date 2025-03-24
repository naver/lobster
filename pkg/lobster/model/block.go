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
	"fmt"
	"time"
)

const (
	TempBlockFileName  = "temp.log"
	blockNameDelimiter = "_"
)

type ReadableBlock interface {
	StartTime() time.Time
	EndTime() time.Time
	FileName() string
	FileNumber() int64
}

type TempBlock struct {
	StartedAt time.Time
	EndedAt   time.Time
	Line      int64
	Size      int64
	FileNum   int64
	IsBackup  bool
}

type Block struct {
	StartedAt    time.Time
	EndedAt      time.Time
	Line         int64
	Size         int64
	FileNum      int64
	DeletionMark bool
}

func NewBlock(start, end time.Time, line, size, fileNum int64) *Block {
	return &Block{
		StartedAt: start,
		EndedAt:   end,
		Line:      line,
		Size:      size,
		FileNum:   fileNum,
	}
}

func NewBlockFromTempBlock(tempBlock TempBlock, fileNum int64) *Block {
	return &Block{
		StartedAt: tempBlock.StartedAt,
		EndedAt:   tempBlock.EndedAt,
		Line:      tempBlock.Line,
		Size:      tempBlock.Size,
		FileNum:   fileNum,
	}
}

func (b Block) StartTime() time.Time {
	return b.StartedAt
}

func (b Block) EndTime() time.Time {
	return b.EndedAt
}

func (b Block) FileName() string {
	return fmt.Sprintf("%s%s%s%s%d%s%d.log",
		b.StartedAt.Local().Format(time.RFC3339Nano),
		blockNameDelimiter,
		b.EndedAt.Local().Format(time.RFC3339Nano),
		blockNameDelimiter,
		b.Line,
		blockNameDelimiter,
		b.FileNum)
}

func (b Block) FileNumber() int64 {
	return b.FileNum
}

func (b TempBlock) FileNumber() int64 {
	return b.FileNum
}

func (b TempBlock) StartTime() time.Time {
	return b.StartedAt
}

func (b TempBlock) EndTime() time.Time {
	return b.EndedAt
}

func (b TempBlock) FileName() string {
	return TempBlockFileName
}

func (b *TempBlock) Reset(fileNum int64) {
	b.StartedAt = time.Time{}
	b.EndedAt = time.Time{}
	b.Line = 0
	b.Size = 0
	b.FileNum = fileNum
}
