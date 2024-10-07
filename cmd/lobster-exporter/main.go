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

package main

import (
	"flag"

	"github.com/naver/lobster/pkg/lobster/logline"
	"github.com/naver/lobster/pkg/lobster/metrics"
	"github.com/naver/lobster/pkg/lobster/server"
	"github.com/naver/lobster/pkg/lobster/sink/exporter"
	"github.com/naver/lobster/pkg/lobster/store"
)

func main() {
	logline.Setup()
	flag.Parse()
	metrics.RegisterMiddlewareMetrics()
	metrics.RegisterSinkMetrics()
	metrics.RegisterExporterMetrics()

	stopChan := make(chan struct{})
	store, err := store.NewStore()
	if err != nil {
		panic(err)
	}
	exporter := exporter.NewLogExporter(store)

	go exporter.Run(stopChan)

	server.NewApiServer(server.Router()).Run(func() {
		close(stopChan)
		store.Clean()
	})
}
