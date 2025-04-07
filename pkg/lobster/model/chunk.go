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
	"os"
	"time"

	"github.com/naver/lobster/pkg/lobster/util"
)

const (
	LabelKeyValueDelimiter = "="
	LabelsDelimiter        = ","
)

// Chunk struct.
type Chunk struct {
	Id                  string      `json:"id"`
	Cluster             string      `json:"cluster"`
	Namespace           string      `json:"namespace"`
	Labels              Labels      `json:"labels"`
	SetName             string      `json:"setName"`
	Pod                 string      `json:"pod"`
	PodUid              string      `json:"podUid"`
	Container           string      `json:"container"`
	Source              Source      `json:"source"`
	Blocks              []*Block    `json:"-"`
	TempBlock           *TempBlock  `json:"-"`
	StartedAt           time.Time   `json:"startedAt"`
	UpdatedAt           time.Time   `json:"updatedAt"`
	DeletionMark        bool        `json:"-"`
	DeletionMarkInBlock bool        `json:"-"`
	Line                int64       `json:"line" format:"int64"`
	Size                int64       `json:"size" format:"int64"`
	CheckPoint          *CheckPoint `json:"-"`
	StoreAddr           string      `json:"storeAddr"`
	RelativePodDir      string      `json:"-"`
	RelativeBlockDir    string      `json:"-"`
}

func NewChunk(file LogFile, checkPoint *CheckPoint) (*Chunk, error) {
	name, err := util.FindSetName(file.Pod)
	if err != nil {
		return nil, err
	}

	return &Chunk{
		Id:               file.RelativeBlockDir(),
		Cluster:          *conf.ClusterName,
		Namespace:        file.Namespace,
		Labels:           file.Labels,
		SetName:          name,
		Pod:              file.Pod,
		PodUid:           file.PodUid,
		Container:        file.Container,
		Source:           file.Source,
		TempBlock:        &TempBlock{},
		StartedAt:        time.Time{},
		UpdatedAt:        time.Time{},
		DeletionMark:     false,
		Line:             0,
		Size:             0,
		CheckPoint:       checkPoint,
		RelativePodDir:   file.RelativePodDir(),
		RelativeBlockDir: file.RelativeBlockDir(),
	}, nil
}

func (c *Chunk) GetBlocksAfterTime(ts time.Time) []ReadableBlock {
	blocks := []ReadableBlock{}
	for _, block := range c.Blocks {
		if block.EndTime().Before(ts) {
			continue
		}
		blocks = append(blocks, block)
	}

	blocks = append(blocks, c.TempBlock)

	return blocks
}

func (c *Chunk) SetCheckPoint(checkPoint *CheckPoint) {
	c.CheckPoint = checkPoint
}

func (c *Chunk) UpdateTempBlock(size, lines int64, ts time.Time) {
	c.TempBlock.Size = c.TempBlock.Size + size
	c.TempBlock.Line = c.TempBlock.Line + lines
	c.TempBlock.EndedAt = ts
	c.UpdatedAt = ts
}

func (c *Chunk) AppendBlocks(blocks []*Block) {
	if c.StartedAt.IsZero() {
		c.StartedAt = blocks[0].StartedAt
	}

	endedAt := blocks[len(blocks)-1].EndedAt
	if endedAt.After(c.UpdatedAt) {
		c.UpdatedAt = endedAt
	}

	line, size := measureBlocks(blocks)
	c.Line = c.Line + line
	c.Size = c.Size + size
	c.Blocks = append(c.Blocks, blocks...)
}

func (c *Chunk) SetTempBlock(block *TempBlock) {
	if c.StartedAt.IsZero() {
		c.StartedAt = block.StartedAt
	}

	if block.EndedAt.After(c.UpdatedAt) {
		c.UpdatedAt = block.EndedAt
	}

	c.Line = c.Line + block.Line
	c.Size = c.Size + block.Size
	c.TempBlock = block
}

func (c *Chunk) MarkBlockAt(i int) {
	c.DeletionMarkInBlock = true
	c.Blocks[i].DeletionMark = true
}

func (c *Chunk) Copy() []*Block {
	copied := make([]*Block, len(c.Blocks))
	copy(copied, c.Blocks)
	return copied
}

func (c Chunk) BlockLength() int {
	return len(c.Blocks)
}

func (c Chunk) LastBlock() *Block {
	if len(c.Blocks) > 0 {
		return c.Blocks[len(c.Blocks)-1]
	}
	return nil
}

func (c Chunk) HasBlocks() bool {
	return c.TempBlock.Size > 0 || len(c.Blocks) > 0
}

func (c Chunk) DeleteContainerFiles(blockPath string) {
	os.RemoveAll(fmt.Sprintf("%s/%s", blockPath, c.RelativeBlockDir))
}

func (c *Chunk) DeleteBlockAt(i int, rootPath string) {
	os.Remove(fmt.Sprintf("%s/%s/%s", rootPath, c.RelativeBlockDir, c.Blocks[i].FileName()))
	if i == 0 && len(c.Blocks) > 1 {
		c.StartedAt = c.Blocks[1].StartedAt
	}
	c.Line = c.Line - c.Blocks[i].Line
	c.Size = c.Size - c.Blocks[i].Size
	c.Blocks = append(c.Blocks[:i], c.Blocks[i+1:]...)
}

func (c Chunk) IsOutdated(retentionTime time.Duration) bool {
	return retentionTime < time.Since(c.UpdatedAt)
}

func (c Chunk) Key() string {
	return fmt.Sprintf("%s_%s_%s_%s_%s_%s", c.Namespace, c.SetName, c.Pod, c.PodUid, c.Container, c.Source.String())
}

func measureBlocks(blocks []*Block) (line, size int64) {
	for _, block := range blocks {
		line = line + block.Line
		size = size + block.Size
	}
	return
}
