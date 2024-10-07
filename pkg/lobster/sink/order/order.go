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

package order

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/query"
	sinkV1 "github.com/naver/lobster/pkg/operator/api/v1"
	v1 "github.com/naver/lobster/pkg/operator/server/api/v1"
)

type Order struct {
	SinkNamespace     string               `json:"sinkNamespace"`
	SinkName          string               `json:"sinkName"`
	SinkType          string               `json:"sinkType"`
	LogMetricRule     sinkV1.LogMetricRule `json:"logMetricRule"`
	LogExportRule     sinkV1.LogExportRule `json:"logExportRule"`
	ContentsNamespace string               `json:"contentsNamespace"`
	ContentsName      string               `json:"contentsName"`
	Request           query.Request        `json:"request"`
}

func NewOrder(sink v1.Sink, sinkContent v1.SinkContents, request query.Request) Order {
	order := Order{
		SinkNamespace:     sink.Namespace,
		SinkName:          sink.Name,
		SinkType:          sink.Type,
		Request:           request,
		ContentsName:      sinkContent.GetName(),
		ContentsNamespace: sinkContent.GetNamespace(),
	}

	switch sink.Type {
	case sinkV1.LogMetricRules:
		order.LogMetricRule = sinkContent.(sinkV1.LogMetricRule)
	case sinkV1.LogExportRules:
		order.LogExportRule = sinkContent.(sinkV1.LogExportRule)
	}

	return order
}

func (o *Order) Clone() (Order, error) {
	var (
		b   bytes.Buffer
		ret Order
	)

	if err := gob.NewEncoder(&b).Encode(o); err != nil {
		return ret, err
	}
	if err := gob.NewDecoder(&b).Decode(&ret); err != nil {
		return ret, err
	}

	return ret, nil
}

func (o *Order) Update(c model.Chunk) error {
	o.Request.Pod = c.Pod
	o.Request.PodUID = c.PodUID
	o.Request.Container = c.Container
	o.Request.Source = c.Source

	return o.Request.InitTextFilterer()
}

func (o Order) Path() string {
	return fmt.Sprintf("%s/%s/%s/%s/%s",
		o.ContentsNamespace,
		o.SinkName,
		o.ContentsName,
		o.Request.Pod,
		o.Request.Container)
}

func (o Order) Key() string {
	return fmt.Sprintf("%s_%s_%s_%s_%s_%s",
		o.ContentsNamespace,
		o.SinkName,
		o.ContentsName,
		o.Request.Pod,
		o.Request.Container,
		o.Request.Source.String())
}
