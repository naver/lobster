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

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/distributor"
	"github.com/naver/lobster/pkg/lobster/logline"
	"github.com/naver/lobster/pkg/lobster/metrics"
	"github.com/naver/lobster/pkg/lobster/proto"
	"github.com/naver/lobster/pkg/lobster/push"
	"github.com/naver/lobster/pkg/lobster/server"
	"github.com/naver/lobster/pkg/lobster/server/handler/log"
	"github.com/naver/lobster/pkg/lobster/server/handler/web"
	"github.com/naver/lobster/pkg/lobster/server/middleware"
	"github.com/naver/lobster/pkg/lobster/store"
)

func main() {
	enableWeb := flag.Bool("enableWeb", false, "Enable web handler")

	logline.Setup()
	flag.Parse()
	metrics.RegisterMiddlewareMetrics()
	metrics.RegisterStoreMetrics()
	metrics.RegisterSinkMetrics()
	metrics.RegisterMatcherMetrics()

	stopChan := make(chan struct{})
	store, err := store.NewStore()
	if err != nil {
		panic(err)
	}

	grpcServer := &proto.ProtoServer{Service: proto.ChunkService{Store: store}}
	distributor := distributor.NewDistributor(store)
	limiter := middleware.Limiter{Limit: store.ReqMaxBurst, CooldownSecond: store.ReqCooldownDuration}
	router := server.Router()

	if *enableWeb {
		webHandler := web.WebHandler{Querier: store}
		router.Handle("/", webHandler)
		router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("/web/static/"))))
	}

	versionedRouter := router.PathPrefix(log.PathApi).Subrouter()
	versionedRouter.Use(middleware.Inspector{}.Middleware)
	versionedRouter.Handle(log.PathLogs, log.ListHandler{Querier: store})
	versionedRouter.Handle(log.PathLogSeries, log.SeriesHandler{Querier: store})
	versionedRouter.Handle(log.PathLogRange, limiter.Middleware(log.RangeHandler{Querier: store, ShouldResponseStringContentsOnly: true}))

	server := server.NewApiServer(router)
	ep := server.GetLocalEndpoint()

	glog.Infof("local endpoint: %s", ep)

	push.Run(store, ep, stopChan)
	distributor.Run(stopChan)
	grpcServer.Run(stopChan)

	server.Run(func() {
		close(stopChan)
		store.Clean()
	})
}
