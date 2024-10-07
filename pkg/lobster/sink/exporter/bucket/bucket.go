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

package bucket

import (
	"fmt"
	"time"

	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/sink/order"
)

const (
	defaultLayout  = "2006-01"
	layoutFileName = time.RFC3339
	timeout        = 10 * time.Second
)

type Bucket interface {
	Flush([]byte, string, string) error
	Interval() time.Duration
	Type() string
	Name() string
	Dir(model.Chunk, time.Time) string
	FileName(time.Time, time.Time) string
	Validate() error
}

func New(order order.Order) (Bucket, error) {
	if order.LogExportRule.S3Bucket != nil {
		return NewS3Bucket(order), nil
	}
	if order.LogExportRule.BasicBucket != nil {
		return NewBasicBucket(order), nil
	}

	return nil, fmt.Errorf("not supported type %s", order.SinkType)
}
