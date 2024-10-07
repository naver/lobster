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
	"time"
)

const BucketPrecision = time.Second

type Bucket struct {
	BucketMetadata
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Lines int64     `json:"lines"`
	Size  uint64    `json:"size"`
}

type BucketMetadata struct {
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Container string `json:"container"`
	Source    Source `json:"source"`
	FileNum   int64  `json:"file_num"`
	ChunkKey  string `json:"chunk_key"`
}

type BucketBuilder struct {
	currentPos BucketMetadata
	buckets    []Bucket
	start      time.Time
	end        time.Time
	lines      int64
	size       uint64
}

func NewBucketBuilder(t time.Time, chunk Chunk) *BucketBuilder {
	return &BucketBuilder{
		currentPos: BucketMetadata{
			Cluster:   chunk.Cluster,
			Namespace: chunk.Namespace,
			Pod:       chunk.Pod,
			Container: chunk.Container,
			Source:    chunk.Source,
			ChunkKey:  chunk.Key(),
			FileNum:   -1,
		},
		start: t.Truncate(BucketPrecision),
		end:   t.Truncate(BucketPrecision).Add(BucketPrecision),
	}
}

func (b *BucketBuilder) Reset(fileNum int64, blockTime time.Time) {
	if b.currentPos.FileNum < 0 {
		b.currentPos.FileNum = fileNum
		return
	}
	if b.currentPos.FileNum != fileNum {
		b.Next(blockTime)
	}
	b.currentPos.FileNum = fileNum
}

func (b BucketBuilder) IsWithinRange(ts time.Time) bool {
	return ts.After(b.start) && ts.Before(b.end)
}

func (b *BucketBuilder) Next(nextTs time.Time) {
	b.Save()
	b.start = nextTs.Truncate(BucketPrecision)
	b.end = nextTs.Truncate(BucketPrecision).Add(BucketPrecision)
	b.lines = 0
	b.size = 0
}

func (b *BucketBuilder) Save() {
	if b.lines == 0 {
		return
	}

	b.buckets = append(b.buckets, Bucket{
		BucketMetadata: b.currentPos,
		Start:          b.start,
		End:            b.end,
		Lines:          b.lines,
		Size:           b.size,
	})
}

func (b *BucketBuilder) Pour(size uint64) {
	b.lines = b.lines + 1
	b.size = b.size + size
}

func (b *BucketBuilder) Build() []Bucket {
	return b.buckets
}
