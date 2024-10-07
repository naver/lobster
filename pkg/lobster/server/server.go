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
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/naver/lobster/pkg/lobster/metrics"
	"github.com/naver/lobster/pkg/lobster/util"
)

var conf config

func init() {
	conf = setup()
	log.Println("server configuration is loaded")
}

type ApiServer struct {
	*http.Server
}

func NewApiServer(router *mux.Router) *ApiServer {
	return &ApiServer{
		&http.Server{
			Addr:         fmt.Sprintf("%s:%s", *conf.ServerAddr, *conf.ServerPort),
			WriteTimeout: *conf.WriteTimeout,
			ReadTimeout:  *conf.ReadTimeout,
			IdleTimeout:  *conf.IdleTimeout,
			Handler:      router,
			ErrorLog:     log.New(os.Stdout, "[SVR_ERR]", log.LstdFlags),
		},
	}
}

func (s *ApiServer) Run(doShutdown func()) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	stopChan := make(chan struct{}, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	s.SetKeepAlivesEnabled(true)
	go func() {
		<-sigs

		close(stopChan)

		if doShutdown != nil {
			doShutdown()
		}
		_ = s.Shutdown(context.Background())
	}()
	if err := metrics.Run(fmt.Sprintf("%s:%s", *conf.MetricsAddr, *conf.MetricsPort), stopChan); err != nil {
		glog.Fatal(err)
	}
	glog.Info("Start server")
	if err := s.ListenAndServe(); err != nil {
		glog.Fatal(err)
	}

}

func (s ApiServer) GetLocalEndpoint() string {
	return fmt.Sprintf("%s:%s", util.GetLocalAddress(), *conf.ServerPort)
}
