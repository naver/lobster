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
	"reflect"

	sinkV1 "github.com/naver/lobster/pkg/operator/api/v1"
)

type SinkContents interface {
	GetNamespace() string
	GetName() string
	GetFilter() sinkV1.Filter
	Validate() error
}

type Sink struct {
	Name           string                 `json:"name,omitempty"`
	Namespace      string                 `json:"namespace,omitempty"`
	Type           string                 `json:"type,omitempty"`
	Description    string                 `json:"description,omitempty"`
	LogMetricRules []sinkV1.LogMetricRule `json:"logMetricRules,omitempty"`
	LogExportRules []sinkV1.LogExportRule `json:"logExportRules,omitempty"`
}

func (s Sink) ListSinkContents() []SinkContents {
	var contentsList []SinkContents

	for _, b := range s.LogExportRules {
		contentsList = append(contentsList, b)
	}

	for _, r := range s.LogMetricRules {
		contentsList = append(contentsList, r)
	}

	return contentsList
}

func (s Sink) Validate() error {
	if len(s.Namespace) == 0 || len(s.Name) == 0 {
		return fmt.Errorf("should set `namespace` and `name`")
	}

	switch s.Type {
	case sinkV1.LogMetricRules:
		return ValidateContent(s.LogMetricRules)
	case sinkV1.LogExportRules:
		return ValidateContent(s.LogExportRules)
	}

	return fmt.Errorf("%s type is not supported", s.Type)
}

func ValidateContent(content interface{}) error {
	existence := map[string]bool{}
	v := reflect.ValueOf(content)

	for i := 0; i < v.Len(); i++ {
		ct := v.Index(i).Interface().(SinkContents)
		name := ct.GetName()
		if err := ct.Validate(); err != nil {
			return err
		}

		if _, ok := existence[name]; ok {
			return fmt.Errorf("duplicated name is not allowed '%s'", name)
		}

		existence[name] = true
	}

	return nil
}

func MergeContent(origin, new interface{}) interface{} {
	existence := map[string]bool{}
	originContent := reflect.ValueOf(origin)
	merged := reflect.ValueOf(new)

	for i := 0; i < merged.Len(); i++ {
		existence[merged.Index(i).Interface().(SinkContents).GetName()] = true
	}

	for i := 0; i < originContent.Len(); i++ {
		item := originContent.Index(i)
		ct := item.Interface().(SinkContents)

		if _, ok := existence[ct.GetName()]; ok {
			continue
		}

		merged = reflect.Append(merged, item)
	}

	return merged.Interface()
}

func SearchContentToDelete(content interface{}, targetName string) int {
	v := reflect.ValueOf(content)

	for i := 0; i < v.Len(); i++ {
		if v.Index(i).Interface().(SinkContents).GetName() != targetName {
			continue
		}
		return i
	}

	return -1
}
