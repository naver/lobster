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
	"sync/atomic"
	"time"
)

type Limiter struct {
	count          int64
	Limit          int64
	CooldownSecond time.Duration
}

func (rl *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loaded := atomic.LoadInt64(&rl.count)

		if loaded < rl.Limit {
			defer func() {
				if loaded+1 == rl.Limit {
					time.Sleep(rl.CooldownSecond)
				}
				atomic.AddInt64(&rl.count, -1)
			}()

			atomic.AddInt64(&rl.count, 1)
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusTooManyRequests)
		}
	})
}
