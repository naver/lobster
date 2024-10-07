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
	"sort"
	"time"
)

type Sample struct {
	Timestamp time.Time `json:"timestamp"`
	Lines     int64     `json:"lines"`
	Size      uint64    `json:"size"`
}

// SeriesData
// @Description Name: "{cluster}_{namespace}_{pod}_{container}_{source}-{file number}".
type Series struct {
	ChunkKey string   `json:"chunk_key"`
	Name     string   `json:"name"`
	Lines    int64    `json:"lines"`
	Size     uint64   `json:"size"`
	Samples  []Sample `json:"samples"`
}

func newSeries(metadata BucketMetadata) *Series {
	return &Series{
		ChunkKey: metadata.ChunkKey,
		Name:     fmt.Sprintf("%s_%s_%s_%s_%s-%d", metadata.Cluster, metadata.Namespace, metadata.Pod, metadata.Container, metadata.Source.String(), metadata.FileNum),
	}
}

func (s *Series) Append(sample Sample) {
	s.Samples = append(s.Samples, sample)
	s.Lines = s.Lines + sample.Lines
	s.Size = s.Size + sample.Size
}

func (s Series) SizeWithinRange(start, end time.Time) (size uint64) {
	for _, sample := range s.Samples {
		if sample.Timestamp.Before(start) || sample.Timestamp.After(end) {
			continue
		}
		size = sample.Size
	}
	return
}

func (s *Series) ReorderSamples() {
	samples := []Sample{}
	sampleMap := map[time.Time]*Sample{}

	for _, s := range s.Samples {
		v, ok := sampleMap[s.Timestamp]
		if !ok {
			sampleMap[s.Timestamp] = &Sample{Timestamp: s.Timestamp, Lines: s.Lines, Size: s.Size}
		} else {
			v.Lines = v.Lines + s.Lines
			v.Size = v.Size + s.Size
		}
	}

	for _, sample := range sampleMap {
		if sample.Lines > 0 {
			samples = append(samples, *sample)
		}
	}

	sort.Slice(samples, func(i, j int) bool {
		return samples[i].Timestamp.Before(samples[j].Timestamp)
	})

	s.Samples = samples
}

// SeriesData
// @Description Array contains Series.
type SeriesData []*Series

func (d SeriesData) Lines() (lines int64) {
	for _, series := range d {
		lines = lines + series.Lines
	}
	return
}

func (d SeriesData) MergedSamples() []Sample {
	samples := []Sample{}
	sampleMap := map[time.Time]*Sample{}

	for _, series := range d {
		for _, s := range series.Samples {
			v, ok := sampleMap[s.Timestamp]
			if !ok {
				sampleMap[s.Timestamp] = &Sample{Timestamp: s.Timestamp, Lines: s.Lines, Size: s.Size}
			} else {
				v.Lines = v.Lines + s.Lines
				v.Size = v.Size + s.Size
			}
		}
	}

	for _, sample := range sampleMap {
		if sample.Lines > 0 {
			samples = append(samples, *sample)
		}
	}

	sort.Slice(samples, func(i, j int) bool {
		return samples[i].Timestamp.Before(samples[j].Timestamp)
	})

	return samples
}

func (d SeriesData) UpdateSamplesByPrecision(precision time.Duration) {
	for i, series := range d {
		newSamples := []Sample{{}}
		index := 0

		for _, sample := range series.Samples {
			if sample.Timestamp.Before(newSamples[index].Timestamp.Add(precision)) {
				newSamples[index].Lines = newSamples[index].Lines + sample.Lines
				newSamples[index].Size = newSamples[index].Size + sample.Size
				continue
			}

			if !newSamples[index].Timestamp.IsZero() {
				newSamples = append(newSamples, Sample{})
				index = index + 1
			}

			newSamples[index].Timestamp = sample.Timestamp.Truncate(precision)
			newSamples[index].Lines = sample.Lines
			newSamples[index].Size = sample.Size
		}

		d[i].Samples = newSamples
	}
}

func BucketsToSeries(start, end time.Time, buckets []Bucket) SeriesData {
	seriesData := []*Series{}
	data := map[string]*Series{}

	for _, bucket := range buckets {
		series := newSeries(bucket.BucketMetadata)

		if _, ok := data[series.Name]; !ok {
			data[series.Name] = series
		}

		data[series.Name].Append(Sample{bucket.Start, bucket.Lines, bucket.Size})
	}

	for _, series := range data {
		series.ReorderSamples()
		seriesData = append(seriesData, series)
	}

	return seriesData
}
