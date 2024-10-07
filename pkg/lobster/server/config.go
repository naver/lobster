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
	"flag"
	"time"
)

type config struct {
	ServerAddr   *string
	ServerPort   *string
	MetricsAddr  *string
	MetricsPort  *string
	WriteTimeout *time.Duration
	ReadTimeout  *time.Duration
	IdleTimeout  *time.Duration
}

func setup() config {
	serverAddr := flag.String("server.addr", "", "server address")
	serverPort := flag.String("server.port", "8880", "server port")
	metricsAddr := flag.String("server.metricsAddr", "", "metrics server address")
	metricsPort := flag.String("server.metricsPort", "8881", "metrics server port")
	writeTimeout := flag.Duration("server.writeTimeout", 300*time.Second, "write timeout seconds")
	readTimeout := flag.Duration("server.readTimeout", 300*time.Second, "read timeout seconds")
	idleTimeout := flag.Duration("server.idleTimeout", 15*time.Second, "idle timeout seconds")

	return config{
		ServerAddr:   serverAddr,
		ServerPort:   serverPort,
		MetricsAddr:  metricsAddr,
		MetricsPort:  metricsPort,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
	}
}
