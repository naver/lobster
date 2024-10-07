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

package log

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/naver/lobster/pkg/lobster/query"
)

const (
	Scheme  = "http"
	PathApi = "/api/{version}"

	ApiV1 = "v1"
	ApiV2 = "v2"
)

var Versions = ApiVersions{ApiV1, ApiV2}

type ApiVersions []string

func (v ApiVersions) IsValid(received string) bool {
	for _, supported := range v {
		if supported == received {
			return true
		}
	}

	return false
}

func parseRequest(r *http.Request) (query.Request, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return query.Request{}, fmt.Errorf("invalid request")
	}

	version, ok := mux.Vars(r)["version"]
	if !ok || !Versions.IsValid(version) {
		return query.Request{}, fmt.Errorf("invalid version")
	}

	req, err := query.ParseRequestWithBody(data)
	if err != nil {
		return query.Request{}, err
	}

	req.Version = version

	return req, nil
}
