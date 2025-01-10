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
	"time"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/metrics"
)

type Recorder struct {
	http.ResponseWriter
	Size   int
	Status int
}

func (r *Recorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *Recorder) Write(data []byte) (int, error) {
	r.Size = len(data)
	return r.ResponseWriter.Write(data)
}

type Inspector struct{}

func (i Inspector) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		measureStart := time.Now()
		recorder := &Recorder{w, 0, 0}

		next.ServeHTTP(recorder, r)

		glog.Infof("[%s] took %fs %s %d %dbytes", r.RemoteAddr, time.Since(measureStart).Seconds(), r.URL.Path, recorder.Status, recorder.Size)

		metrics.ObserveHandleSeconds(r.URL.Path, time.Since(measureStart).Seconds())
		metrics.AddResponseBytes(r.URL.Path, recorder.Size)
		metrics.AddResponseStatus(r.URL.Path, correctStatus(recorder.Status))
	})
}

func correctStatus(status int) int {
	if status == 0 {
		return http.StatusOK
	}

	return status
}
