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
	"os"

	"github.com/naver/lobster/pkg/lobster/query"
)

type StdstreamStub struct{}

func (s StdstreamStub) Keyword() string {
	return "stdstream"
}

func (s StdstreamStub) GenerateLogs(conf Config, stopChan chan struct{}) {
	generateLogs(log.New(os.Stdout, s.Keyword(), 0), conf, stopChan, func(log string) string {
		return log
	})
}

func (s StdstreamStub) Query() query.Request {
	namespace := os.Getenv("NAMESPACE")
	pod := os.Getenv("POD")
	container := os.Getenv("CONTAINER")

	return query.Request{
		ID:                "loggen-test-stdstream",
		Namespaces:        []string{namespace},
		Pods:              []string{pod},
		Containers:        []string{container},
		Page:              -1,
		FilterIncludeExpr: s.Keyword(),
	}
}
