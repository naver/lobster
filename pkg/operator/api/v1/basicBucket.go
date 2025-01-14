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
	"github.com/naver/lobster/pkg/operator/api/v1/template"
)

type BasicBucket struct {
	// Address to export logs
	Destination string `json:"destination,omitempty"`
	// Deprecated; Root directory to store logs within external storage
	RootPath string `json:"rootPath,omitempty"`
	// Deprecated; An option(default `2006-01`) that sets the name of the sub-directory following `{Root path}` to a time-based layout
	TimeLayoutOfSubDirectory string `json:"timeLayoutOfSubDirectory,omitempty" default:"2006-01"`
	// Provide an option to convert '+' to '%2B' to address issues in certain web environments where '+' is misinterpreted
	ShouldEncodeFileName bool `json:"shouldEncodeFileName,omitempty"`
	// Path constructed from log metadata for exporting logs
	PathTemplate string `json:"pathTemplate,omitempty"`
}

func (b BasicBucket) Validate() ValidationErrors {
	var validationErrors ValidationErrors

	if len(b.Destination) == 0 {
		validationErrors.AppendErrorWithFields("basicBucket.destination", ErrorEmptyField)
	}

	if err := template.ValidateTemplateString(b.PathTemplate); err != nil {
		validationErrors.AppendErrorWithFields("basicBucket.pathTemplate", err.Error())
	}

	return validationErrors
}
