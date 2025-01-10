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

	"github.com/naver/lobster/pkg/lobster/global"
	"github.com/naver/lobster/pkg/lobster/logline"
	"github.com/naver/lobster/pkg/lobster/metrics"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/naver/lobster/pkg/lobster/server"

	"github.com/naver/lobster/pkg/lobster/server/handler/log"
	"github.com/naver/lobster/pkg/lobster/server/handler/web"
	"github.com/naver/lobster/pkg/lobster/server/middleware"

	_ "github.com/naver/lobster/pkg/docs/global-query"
)

//	@title			Lobster API document
//	@version		1.0
//	@description	Descriptions of Lobster global query APIs

func main() {
	logline.Setup()
	flag.Parse()
	metrics.RegisterMiddlewareMetrics()
	metrics.RegisterQuerierMetrics()

	stopChan := make(chan struct{})
	querier := global.NewQuerier()

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

	server := server.NewApiServer(router)

	server.Run(func() {
		close(stopChan)
	})
}
