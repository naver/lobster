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
	"strings"

	"github.com/naver/lobster/pkg/operator/api/v1/template"
)

const MaxS3Tags = 10

type Tags map[string]string

// The tag-set must be encoded as URL Query parameters.
func (t Tags) String() string {
	var str []string

	for k, v := range t {
		str = append(str, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(str, "&")
}

type S3Bucket struct {
	// S3 Address to export logs
	Destination string `json:"destination,omitempty"`
	// Deprecated; Root directory to store logs within external storage
	RootPath string `json:"rootPath,omitempty"`
	// Deprecated; An option(default `2006-01`) that sets the name of the sub-directory following `{Root path}` to a time-based layout
	TimeLayoutOfSubDirectory string `json:"timeLayoutOfSubDirectory,omitempty" default:"2006-01"`
	// S3 bucket name
	BucketName string `json:"bucketName,omitempty"`
	// S3 region
	Region string `json:"region,omitempty"`
	// S3 bucket access key
	AccessKey string `json:"accessKey,omitempty"`
	// S3 bucket secret key
	SecretKey string `json:"secretKey,omitempty"`
	// Tags for objects to be stored
	Tags Tags `json:"tags,omitempty"`
	// Provide an option to convert '+' to '%2B' to address issues in certain web environments where '+' is misinterpreted
	ShouldEncodeFileName bool `json:"shouldEncodeFileName,omitempty"`
	// Path constructed from log metadata for exporting logs
	PathTemplate string `json:"pathTemplate,omitempty"`
}

func (s S3Bucket) Validate() error {
	if len(s.Destination) == 0 || len(s.RootPath) == 0 {
		return fmt.Errorf("`destination` and `rootPath` should not be empty")
	}

	if len(s.AccessKey) == 0 || len(s.SecretKey) == 0 {
		return fmt.Errorf("`accessKey` and `secretKey` should not be empty")
	}

	if MaxS3Tags < len(s.Tags) {
		return fmt.Errorf("too many tags")
	}

	if err := template.ValidateTemplateString(s.PathTemplate); err != nil {
		return err
	}

	return nil
}
