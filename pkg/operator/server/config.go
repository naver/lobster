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
	Addr         string
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
	IdleTimeout  time.Duration
	MaxSinkRule  int
}

func Setup() {
	flag.StringVar(&conf.Addr, "addr", ":8080", "server address")
	flag.IntVar(&conf.MaxSinkRule, "maxSinkRule", 50, "maximum number of sink rules")
	flag.DurationVar(&conf.WriteTimeout, "writeTimeout", 10*time.Second, "server write timeout")
	flag.DurationVar(&conf.ReadTimeout, "readTimeout", 10*time.Second, "server read timeout")
	flag.DurationVar(&conf.IdleTimeout, "idleTimeout", 10*time.Second, "server idle timeout")
}
