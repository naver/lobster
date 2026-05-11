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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"

	sinkV1 "github.com/naver/lobster/pkg/operator/api/v1"
)

var invalidNameCharacter = regexp.MustCompile(`[<>:"/\\|?*]`)

type SinkRule interface {
	GetNamespace() string
	GetName() string
	GetFilter() sinkV1.Filter
	Validate() sinkV1.ValidationErrors
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

func (s Sink) Validate() sinkV1.ValidationErrors {
	var validationErrors sinkV1.ValidationErrors

	if len(s.Namespace) == 0 {
		validationErrors.AppendErrorWithFields("lobsterSink.namespace", sinkV1.ErrorEmptyField)
	}

	if len(s.Namespace) == 0 {
		validationErrors.AppendErrorWithFields("lobsterSink.name", sinkV1.ErrorEmptyField)
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

func ValidateRules(rules any) sinkV1.ValidationErrors {
	existence := map[string]bool{}
	v := reflect.ValueOf(rules)

	for i := 0; i < v.Len(); i++ {
		rule := v.Index(i).Interface().(SinkRule)
		name := rule.GetName()
		if errList := rule.Validate(); !errList.IsEmpty() {
			return errList
		}

		if _, ok := existence[name]; ok {
			return sinkV1.ValidationErrors{sinkV1.NewValidationError("{logMetricRules|logExportRules}.name", fmt.Sprintf("duplicated name is not allowed '%s'", name))}
		}

		if err := hasValidName(name); err != nil {
			return sinkV1.ValidationErrors{sinkV1.NewValidationError("{logMetricRules|logExportRules}.name", err.Error())}
		}

		existence[name] = true
	}

	return nil
}

func MergeRules(origin, new any) any {
	originRules := reflect.ValueOf(origin)
	merged := reflect.ValueOf(new)

	originByName := make(map[string]any, originRules.Len())
	for i := 0; i < originRules.Len(); i++ {
		rule := originRules.Index(i).Interface()
		originByName[rule.(SinkRule).GetName()] = rule
	}

	inNew := make(map[string]bool, merged.Len())
	for i := 0; i < merged.Len(); i++ {
		newRule := merged.Index(i).Interface()
		name := newRule.(SinkRule).GetName()
		inNew[name] = true
		if orig, ok := originByName[name]; ok {
			merged.Index(i).Set(reflect.ValueOf(overlayFields(orig, newRule)))
		}
	}

	for i := 0; i < originRules.Len(); i++ {
		item := originRules.Index(i)
		if !inNew[item.Interface().(SinkRule).GetName()] {
			merged = reflect.Append(merged, item)
		}
	}

	return merged.Interface()
}

// overlayFields returns a copy of origin with non-zero fields from new applied on top.
// It relies on omitempty JSON tags: marshaling new omits zero-value fields,
// so only explicitly set fields in new overwrite the corresponding fields in origin.
func overlayFields(origin, new any) any {
	b, _ := json.Marshal(new)
	result := reflect.New(reflect.TypeOf(origin))
	result.Elem().Set(reflect.ValueOf(origin))
	_ = json.Unmarshal(b, result.Interface())
	return result.Elem().Interface()
}

func SearchRuleToDelete(rule any, targetName string) int {
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
