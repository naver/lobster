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

package web

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/naver/lobster/pkg/lobster/model"
)

const (
	scriptTemplate = `
    var chart = bb.generate({
	data: {
	  x: "date",
	  columns: [
		{{ .Labels }}
		{{ .Values }}
	  ],
	type: "bar",
	},
	axis: {
	  x: {
		type: "timeseries",
		tick: {
		  format: "%Y-%m-%d %H:%M:%S"
		}
	  }
	},
	bindto: "#historgram"
  });`
	EmptySampleValue = 0
	precision        = time.Minute
)

var tScript *template.Template

func init() {
	tScript = template.Must(template.New("script").Parse(scriptTemplate))
}

func executeTemplate(seriesData model.SeriesData) (string, error) {
	var buffer bytes.Buffer
	var labels string
	var values string

	if len(seriesData) > 0 {
		seriesData.UpdateSamplesByPrecision(precision)
		timestamps := timestampsInOrder(seriesData)
		labels = makeLabel(timestamps)
		values = makeValue(seriesData, timestamps)
	}

	err := tScript.Execute(&buffer, struct {
		Labels string
		Values string
	}{labels, values})

	return buffer.String(), err
}

func makeLabel(timestamps []time.Time) string {
	data := []string{"\"date\""}

	for _, t := range timestamps {
		data = append(data, fmt.Sprintf("\"%s\"", t.Format("2006-01-02 15:04:05")))
	}

	return fmt.Sprintf("[%s],", strings.Join(data, ","))
}

func makeValue(seriesData model.SeriesData, timestamps []time.Time) string {
	data := []string{}

	for _, series := range seriesData {
		data = append(data, formatData(series.Name, fillSampleValueByTimestamps(series.Samples, timestamps)))
	}

	return strings.Join(data, ",")
}

func fillSampleValueByTimestamps(samples []model.Sample, timestamps []time.Time) []int64 {
	sampleValues := []int64{}
	tsIndex := 0

	for _, sample := range samples {
		isEmptyFilled := false
		for tsIndex < len(timestamps) && !sample.Timestamp.Equal(timestamps[tsIndex]) {
			tsIndex = tsIndex + 1
			sampleValues = append(sampleValues, EmptySampleValue)
			isEmptyFilled = true
		}

		if isEmptyFilled {
			sampleValues[len(sampleValues)-1] = sample.Lines
		} else {
			sampleValues = append(sampleValues, sample.Lines)
		}
	}

	return sampleValues
}

func formatData(seriesName string, sampleValues []int64) string {
	data := append([]string{fmt.Sprintf("\"%s\"", seriesName)},
		strings.ReplaceAll(strings.Trim(fmt.Sprint(sampleValues), "[]"), " ", ","))
	return fmt.Sprintf("[%s]", strings.Join(data, ","))
}

func timestampsInOrder(seriesData model.SeriesData) []time.Time {
	results := []time.Time{}
	timestamps := []time.Time{}

	for _, series := range seriesData {
		for _, sample := range series.Samples {
			timestamps = append(timestamps, sample.Timestamp)
		}
	}

	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i].Before(timestamps[j])
	})

	prevTime := time.Time{}

	for _, t := range timestamps {
		if !t.Equal(prevTime) {
			results = append(results, t)
		}
		prevTime = t
	}

	return results
}
