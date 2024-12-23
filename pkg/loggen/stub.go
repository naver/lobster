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
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/naver/lobster/pkg/lobster/query"
)

type Stub interface {
	GenerateLogs(Config, chan struct{})
	Query() query.Request
	Keyword() string
}

func generateLogs(logger *log.Logger, conf Config, stopChan chan struct{}, logFormatter func(string) string) {
	ticker := time.NewTicker(*conf.LogGenerationInterval)
	index := uint(0)
	logSequences := []string{}

	for i := 0; i < 10; i++ {
		var str strings.Builder

		for str.Len() < *conf.LogSize {
			str.WriteString(fmt.Sprintf("%d", i))
		}

		logSequences = append(logSequences, str.String())
	}

	for {
		select {
		case <-ticker.C:
			index = (index + 1) % 10

			for i := 0; i < *conf.LogLines; i++ {
				logger.Print(logFormatter(logSequences[index]))
			}

		case <-stopChan:
			return
		}
	}
}
