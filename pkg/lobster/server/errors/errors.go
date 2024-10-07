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

package errors

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/golang/glog"
)

var (
	ErrTooManyRequests     = errors.New("too many requests")
	ErrNotImplemented      = errors.New("not implemented")
	ErrBadRequest          = errors.New("bad request")
	ErrInternalServerError = errors.New("internal error")
)

func ErrorByStatusCode(statusCode int) error {
	switch statusCode {
	case http.StatusTooManyRequests:
		return ErrTooManyRequests
	case http.StatusNotImplemented:
		return ErrNotImplemented
	case http.StatusBadRequest:
		return ErrBadRequest
	}

	return ErrInternalServerError
}

func HandleError(w http.ResponseWriter, err error) {
	switch errors.Cause(err) {
	case ErrTooManyRequests:
		http.Error(w, err.Error(), http.StatusTooManyRequests)
	case ErrNotImplemented:
		http.Error(w, err.Error(), http.StatusNotImplemented)
	case ErrBadRequest:
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		glog.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
