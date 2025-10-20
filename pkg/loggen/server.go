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

package loggen

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const serverTimeout = 10 * time.Second

func RunServer(serverAddr string, stopChan chan struct{}) {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	svr := &http.Server{
		Addr:         serverAddr,
		WriteTimeout: serverTimeout,
		ReadTimeout:  serverTimeout,
		IdleTimeout:  serverTimeout,
		Handler:      router,
		ErrorLog:     log.New(os.Stdout, "[SVR_ERR]", log.LstdFlags),
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
}
