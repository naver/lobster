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

package querier

import (
	"bufio"
	"bytes"
	"io"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/golang/glog"

	"github.com/naver/lobster/pkg/lobster/logline"
	"github.com/naver/lobster/pkg/lobster/model"
)

type ParseFunc func(string, model.Chunk) (model.Entry, error)

func ParseEntry(line string, chunk model.Chunk) (model.Entry, error) {
	var (
		e   = model.NewEntry(chunk)
		err error
	)

	e.Timestamp, err = logline.ParseTimestamp(line)
	if err != nil {
		return e, err
	}

	if chunk.Source.Type == model.LogTypeStdStream {
		e.Stream, err = logline.ParseStream(line)
		if err != nil {
			return e, err
		}

		e.Tag, err = logline.ParseTag(line)
		if err != nil {
			return e, err
		}

		e.Message, err = logline.ParseLogMessage(line)
		if err != nil {
			return e, err
		}
	} else {
		e.Message = line
	}

	e.Message = stripNewline(e.Message)

	return e, nil
}

func ParseEntryRaw(line string, chunk model.Chunk) (model.Entry, error) {
	var (
		e   = model.NewEntry(chunk)
		err error
	)

	e.Timestamp, err = logline.ParseTimestamp(line)
	if err != nil {
		return e, err
	}

	e.Message = line

	return e, nil
}

func stripNewline(msg string) string {
	if len(msg) == 0 {
		return msg
	}

	return msg[:len(msg)-1]
}

func (b *SeriesBuilder) Build() model.SeriesData {
	return b.seriesData
}

type EntryBuilder struct {
	FetchResults []FetchResult
	entries      []model.Entry
	seriesData   model.SeriesData
	total        atomic.Uint64
	limit        uint64
}

func NewEntryBuilder(FetchResults []FetchResult, limit uint64) *EntryBuilder {
	return &EntryBuilder{
		FetchResults: FetchResults,
		entries:      []model.Entry{},
		seriesData:   model.SeriesData{},
		limit:        limit,
	}
}

func (b *EntryBuilder) Merge(fn ParseFunc) *EntryBuilder {
	channel := make(chan []model.Entry)

	for _, r := range b.FetchResults {
		go func(r FetchResult) {
			reader := bufio.NewReader(strings.NewReader(r.response.Contents))
			entries := []model.Entry{}

			for {
				if b.total.Load() > b.limit {
					break
				}
				line, err := reader.ReadString('\n')
				if err != nil {
					if err != io.EOF {
						glog.Error(err)
					}
					break
				}

				e, err := fn(line, r.Chunk)
				if err != nil {
					glog.Error(err)
					continue
				}

				entries = append(entries, e)
				b.total.Add(uint64(len(e.Message)))
			}

			channel <- entries
		}(r)
	}

	for i := 0; i < len(b.FetchResults); i++ {
		entries := <-channel
		b.entries = append(b.entries, entries...)
	}

	return b
}

func (b *EntryBuilder) SortAscending() *EntryBuilder {
	sort.Slice(b.entries, func(i, j int) bool {
		return b.entries[i].Timestamp.Before(b.entries[j].Timestamp)
	})
	return b
}

func (b *EntryBuilder) Build() ([]model.Entry, bool) {
	return b.entries, b.isPartialContents()
}

func (b *EntryBuilder) BuildRawLogs() ([]byte, bool) {
	out := &bytes.Buffer{}

	for _, e := range b.entries {
		out.WriteString(e.String())
	}

	return out.Bytes(), b.isPartialContents()
}

func (b *EntryBuilder) BuildWithLimit(limit int) ([]byte, model.SeriesData, bool) {
	out := &bytes.Buffer{}

	for _, e := range b.entries {
		if limit < out.Len()+len(e.String()) {
			return out.Bytes(), b.seriesData, true
		}
		out.WriteString(e.String())
	}

	return out.Bytes(), b.seriesData, false
}

func (b *EntryBuilder) Len() int {
	return len(b.entries)
}

func (b *EntryBuilder) isPartialContents() bool {
	return b.total.Load() >= b.limit
}

type SeriesBuilder struct {
	FetchResults []FetchResult
	seriesData   model.SeriesData
}

func NewSeriesBuilder(FetchResults []FetchResult) *SeriesBuilder {
	return &SeriesBuilder{
		FetchResults: FetchResults,
		seriesData:   model.SeriesData{},
	}
}

func (b *SeriesBuilder) Merge() *SeriesBuilder {
	for _, r := range b.FetchResults {
		if r.response.SeriesData == nil {
			continue
		}

		b.seriesData = append(b.seriesData, *r.response.SeriesData...)
	}

	return b
}
