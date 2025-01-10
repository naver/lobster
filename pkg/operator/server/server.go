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

package server

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"sigs.k8s.io/controller-runtime/pkg/client"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/naver/lobster/pkg/lobster/server/middleware"
	"github.com/naver/lobster/pkg/operator/server/controller"
	"github.com/naver/lobster/pkg/operator/server/handler"

	_ "net/http/pprof"
)

//	@title			Lobster Operator APIs document
//	@version		1.0
//	@description	Descriptions of Lobster log-sink management APIs

var conf = &config{}

func Run(sinkClient client.Client, logger logr.Logger) {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	stopChan := make(chan struct{})
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{
		Addr:         conf.Addr,
		WriteTimeout: conf.WriteTimeout,
		ReadTimeout:  conf.ReadTimeout,
		IdleTimeout:  conf.IdleTimeout,
		Handler:      setupRouter(sinkClient, logger),
		ErrorLog:     log.New(os.Stdout, "[SVR_ERR]", log.LstdFlags),
	}

	server.SetKeepAlivesEnabled(true)
	go func() {
		<-sigs

		close(stopChan)
		_ = server.Shutdown(context.Background())
	}()

	logger.Info("Start server")
	if err := server.ListenAndServe(); err != nil {
		logger.Error(err, "server error")
	}
}

func setupRouter(sinkClient client.Client, logger logr.Logger) *mux.Router {
	router := mux.NewRouter()
	router.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("/web/static/"))))
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/static/docs/swagger.json"),
	))

	ctrl := controller.SinkController{Client: sinkClient, MaxContent: conf.MaxContent, Logger: logger}
	router.Handle(handler.PathSync, handler.InternalSyncHandler{Ctrl: ctrl, Logger: logger})

	routerV1 := router.PathPrefix(handler.PathApi).Subrouter()
	routerV1.Use(middleware.Inspector{}.Middleware)
	routerV1.Handle(handler.PathSinks, handler.SinkHandler{Ctrl: ctrl, Logger: logger})
	routerV1.Handle(handler.PathSpecificSink, handler.SinkHandler{Ctrl: ctrl, Logger: logger})
	routerV1.Handle(handler.PathSinkContentsRule, handler.SinkHandler{Ctrl: ctrl, Logger: logger}).Methods(http.MethodDelete)

	return router
}
