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

package metrics

import (
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	labelTargetNamespace = "target_namespace"

	labelSinkName         = "sink_name"
	labelSinkNamespace    = "sink_namespace"
	labelSinkType         = "sink_type"
	labelSinkContentsName = "sink_contents_name"

	labelLogNamespace  = "log_namespace"
	labelLogPod        = "log_pod"
	labelLogContainer  = "log_container"
	labelLogSourceType = "log_source_type"
	labelLogSourcePath = "log_source_path"

	labelHandler    = "handler"
	labelStatusCode = "code"
	labelLimit      = "limit"

	metricPath = "/metrics"
)

const serverTimeout = 10 * time.Second

var mutex sync.Mutex

func Run(serverAddr string, stopChan chan struct{}) error {
	mux := http.NewServeMux()
	mux.Handle(metricPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		promhttp.Handler().ServeHTTP(w, r)
	}))

	svr := &http.Server{
		Addr:         serverAddr,
		WriteTimeout: serverTimeout,
		ReadTimeout:  serverTimeout,
		IdleTimeout:  serverTimeout,
		Handler:      mux,
		ErrorLog:     log.New(os.Stdout, "[METRICS_SVR_ERR]", log.LstdFlags),
	}

	go func() {
		select {
		case <-stopChan:
			if err := svr.Close(); err != nil {
				glog.Error(err)
			}
			return
		default:
			if err := svr.ListenAndServe(); err != nil {
				svr.ErrorLog.Fatal(err)
			}
		}

	}()

	return nil
}

func promLabelsKeys(labels prometheus.Labels) []string {
	keys := []string{}

	for k := range labels {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}
