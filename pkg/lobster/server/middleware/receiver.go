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

package middleware

import (
	"net/http"

	"github.com/naver/lobster/pkg/lobster/hash"
)

var headerRealIp = http.CanonicalHeaderKey("X-Real-IP")

type Receiver struct {
	Id       uint64
	Operator hash.HashOperator
}

func (rc Receiver) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rc.Id == rc.Operator.Modulo(r.Header.Get(headerRealIp)) {
			next.ServeHTTP(w, r)
		}
	})
}
