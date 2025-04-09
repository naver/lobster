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
	"errors"
	"fmt"
	"reflect"
	"regexp"

	sinkV1 "github.com/naver/lobster/pkg/operator/api/v1"
	v1 "github.com/naver/lobster/pkg/operator/api/v1"
)

var invalidNameCharacter = regexp.MustCompile(`[<>:"/\\|?*]`)

type SinkRule interface {
	GetNamespace() string
	GetName() string
	GetFilter() sinkV1.Filter
	Validate() v1.ValidationErrors
}

type Sink struct {
	Name           string                 `json:"name,omitempty"`
	Namespace      string                 `json:"namespace,omitempty"`
	Type           string                 `json:"type,omitempty"`
	Description    string                 `json:"description,omitempty"`
	LogMetricRules []sinkV1.LogMetricRule `json:"logMetricRules,omitempty"`
	LogExportRules []sinkV1.LogExportRule `json:"logExportRules,omitempty"`
}

func (s Sink) ListSinkRules() []SinkRule {
	var rules []SinkRule

	for _, b := range s.LogExportRules {
		rules = append(rules, b)
	}

	for _, r := range s.LogMetricRules {
		rules = append(rules, r)
	}

	return rules
}

func (s Sink) Validate() v1.ValidationErrors {
	var validationErrors v1.ValidationErrors

	if len(s.Namespace) == 0 {
		validationErrors.AppendErrorWithFields("lobsterSink.namespace", v1.ErrorEmptyField)
	}

	if len(s.Namespace) == 0 {
		validationErrors.AppendErrorWithFields("lobsterSink.name", v1.ErrorEmptyField)
	}

	switch s.Type {
	case sinkV1.LogMetricRules:
		if errList := ValidateRules(s.LogMetricRules); !errList.IsEmpty() {
			validationErrors.AppendErrors(errList...)
		}
	case sinkV1.LogExportRules:
		if errList := ValidateRules(s.LogExportRules); !errList.IsEmpty() {
			validationErrors.AppendErrors(errList...)
		}
	default:
		validationErrors.AppendErrorWithFields("lobsterSink.type", "unsupported lobsterSink type")

	}

	return validationErrors
}

func ValidateRules(rules interface{}) v1.ValidationErrors {
	existence := map[string]bool{}
	v := reflect.ValueOf(rules)

	for i := 0; i < v.Len(); i++ {
		rule := v.Index(i).Interface().(SinkRule)
		name := rule.GetName()
		if errList := rule.Validate(); !errList.IsEmpty() {
			return errList
		}

		if _, ok := existence[name]; ok {
			return v1.ValidationErrors{v1.NewValidationError("{logMetricRules|logExportRules}.name", fmt.Sprintf("duplicated name is not allowed '%s'", name))}
		}

		if err := hasValidName(name); err != nil {
			return v1.ValidationErrors{v1.NewValidationError("{logMetricRules|logExportRules}.name", err.Error())}
		}

		existence[name] = true
	}

	return nil
}

func MergeRules(origin, new interface{}) interface{} {
	existence := map[string]bool{}
	originRules := reflect.ValueOf(origin)
	merged := reflect.ValueOf(new)

	for i := 0; i < merged.Len(); i++ {
		existence[merged.Index(i).Interface().(SinkRule).GetName()] = true
	}

	for i := 0; i < originRules.Len(); i++ {
		item := originRules.Index(i)
		ct := item.Interface().(SinkRule)

		if _, ok := existence[ct.GetName()]; ok {
			continue
		}

		merged = reflect.Append(merged, item)
	}

	return merged.Interface()
}

func SearchRuleToDelete(rule interface{}, targetName string) int {
	v := reflect.ValueOf(rule)

	for i := 0; i < v.Len(); i++ {
		if v.Index(i).Interface().(SinkRule).GetName() != targetName {
			continue
		}
		return i
	}

	return -1
}

func hasValidName(name string) error {
	if invalidNameCharacter.MatchString(name) {
		return errors.New("invalid characters(<>:\"/\\) are included in name")
	}

	return nil
}
