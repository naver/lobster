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
	MinBucketInterval = 5 * time.Second
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
	// Generate metrics from logs using target or log-based rules
	Filter Filter `json:"filter,omitempty"`
	// Interval to export logs
	Interval metav1.Duration `json:"interval,omitempty" swaggertype:"string" example:"time duration(e.g. 1m)"`
}

func (r LogExportRule) Validate() ValidationErrors {
	var validationErrors ValidationErrors

	if len(r.Name) == 0 {
		validationErrors.AppendErrorWithFields("logExportRule.name", ErrorEmptyField)
	}

	if r.Interval.Seconds() < MinBucketInterval.Seconds() {
		validationErrors.AppendErrorWithFields("logExportRule.interval",
			fmt.Sprintf("`interval` should be greater than or equal to %dm", int(MinBucketInterval.Seconds())))
	} else if MaxBucketInterval.Seconds() < r.Interval.Seconds() {
		validationErrors.AppendErrorWithFields("logExportRule.interval",
			fmt.Sprintf("`interval` should be less than or equal to %dh", int(MaxBucketInterval.Hours())))
	}

	if r.BasicBucket != nil {
		if errList := r.BasicBucket.Validate(); !errList.IsEmpty() {
			validationErrors.AppendErrors(errList...)
		}
	}

	if r.S3Bucket != nil {
		if errList := r.S3Bucket.Validate(); !errList.IsEmpty() {
			validationErrors.AppendErrors(errList...)
		}
	}

	if r.Kafka != nil {
		if errList := r.Kafka.Validate(); !errList.IsEmpty() {
			validationErrors.AppendErrors(errList...)
		}
	}

	if errList := r.Filter.Validate(); !errList.IsEmpty() {
		validationErrors.AppendErrors(errList...)
	}

	return validationErrors
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
