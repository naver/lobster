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

package v1

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	MinBucketInterval = 15 * time.Second
	MaxBucketInterval = time.Hour
)

type LogExportRule struct {
	// Rule name
	Name string `json:"name,omitempty"`
	// Description of this rule
	Description string `json:"description,omitempty"`
	// Settings required to export logs to basic bucket
	BasicBucket *BasicBucket `json:"basicBucket,omitempty"`
	// Settings required to export logs to S3 bucket
	S3Bucket *S3Bucket `json:"s3Bucket,omitempty"`
	// Settings required to export logs to Kafka
	Kafka *Kafka `json:"kafka,omitempty"`
	// Generate metrics from logs by specifying the target or log content
	Filter Filter `json:"filter,omitempty"`
	// Interval to export logs
	Interval metav1.Duration `json:"interval,omitempty" swaggertype:"string" example:"time duration(e.g. 1m)"`
}

func (r LogExportRule) Validate() error {
	if len(r.Name) == 0 {
		return fmt.Errorf("`name` should not be empty")
	}

	if r.Interval.Seconds() < MinBucketInterval.Seconds() {
		return fmt.Errorf("`interval` should be greater than or equal to %dm", int(MinBucketInterval.Seconds()))
	}

	if MaxBucketInterval.Seconds() < r.Interval.Seconds() {
		return fmt.Errorf("`interval` should be less than or equal to %dh", int(MaxBucketInterval.Hours()))
	}

	if r.BasicBucket != nil {
		if err := r.BasicBucket.Validate(); err != nil {
			return err
		}
	}

	if r.S3Bucket != nil {
		if err := r.S3Bucket.Validate(); err != nil {
			return err
		}
	}

	return r.Filter.Validate()
}

func (r LogExportRule) GetName() string {
	return r.Name
}

func (r LogExportRule) GetNamespace() string {
	return r.Filter.Namespace
}

func (r LogExportRule) GetFilter() Filter {
	return r.Filter
}
