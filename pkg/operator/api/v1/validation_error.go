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

import "encoding/json"

const (
	ErrorEmptyField = "this field must not be empty"
)

type ValidationErrors []ValidationError

func (ve ValidationErrors) IsEmpty() bool {
	return len(ve) == 0
}

func (ve *ValidationErrors) AppendErrors(e ...ValidationError) {
	*ve = append(*ve, e...)
}

func (ve *ValidationErrors) AppendErrorWithFields(field, message string) {
	*ve = append(*ve, NewValidationError(field, message))
}

func (e ValidationErrors) String() string {
	jsonData, _ := json.Marshal(e)
	return string(jsonData)
}

type ValidationError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message,omitempty"`
}

func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}
