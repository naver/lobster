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

type LogMetricRule struct {
	// Rule name
	Name string `json:"name,omitempty"`
	// Description of this rule
	Description string `json:"description,omitempty"`
	// Generate metrics from logs using target or log-based rules
	Filter Filter `json:"filter,omitempty"`
}

func (r LogMetricRule) Validate() ValidationErrors {
	var validationErrors ValidationErrors

	if len(r.Name) == 0 {
		validationErrors.AppendErrorWithFields("logMetricRule.name", ErrorEmptyField)
	}

	if errList := r.Filter.Validate(); !errList.IsEmpty() {
		validationErrors.AppendErrors(errList...)
	}

	return validationErrors
}

func (r LogMetricRule) GetName() string {
	return r.Name
}

func (r LogMetricRule) GetNamespace() string {
	return r.Filter.Namespace
}

func (r LogMetricRule) GetFilter() Filter {
	return r.Filter
}
