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
	"net/http"

	"github.com/naver/lobster/pkg/lobster/hash"
	"github.com/naver/lobster/pkg/lobster/logline"
	"github.com/naver/lobster/pkg/lobster/metrics"
	"github.com/naver/lobster/pkg/lobster/querier"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/naver/lobster/pkg/lobster/server"
	"github.com/naver/lobster/pkg/lobster/server/handler/log"
	"github.com/naver/lobster/pkg/lobster/server/handler/push"
	"github.com/naver/lobster/pkg/lobster/server/handler/web"
	"github.com/naver/lobster/pkg/lobster/server/middleware"

	_ "github.com/naver/lobster/pkg/docs/query"
)

//	@title			Lobster API document
//	@version		1.0
//	@description	Descriptions of Lobster query APIs

func main() {
	logline.Setup()
	flag.Parse()
	metrics.RegisterMiddlewareMetrics()
	metrics.RegisterQuerierMetrics()

	stopChan := make(chan struct{})
	querier := querier.NewQuerier()

	receiver := middleware.Receiver{Id: querier.Id, Operator: hash.HashOperator{Modulus: querier.Modulus}}
	router := server.Router()

	webHandler := web.WebHandler{Querier: querier}
	router.Handle("/", webHandler)

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("/web/static/"))))
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/static/docs/swagger.json"),
	))

	versionedRouter := router.PathPrefix(log.PathApi).Subrouter()
	versionedRouter.Use(middleware.Inspector{}.Middleware)
	versionedRouter.Handle(log.PathLogs, log.ListHandler{Querier: querier})
	versionedRouter.Handle(log.PathLogSeries, log.SeriesHandler{Querier: querier})
	versionedRouter.Handle(log.PathLogRange, log.RangeHandler{Querier: querier})

	router.Handle(push.PathPush, receiver.Middleware(push.PushHandler{Querier: querier}))

	server := server.NewApiServer(router)

	querier.Run(stopChan)
	server.Run(func() {
		close(stopChan)
	})
}
