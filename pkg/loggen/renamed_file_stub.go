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
	"os"
	"syscall"
	"time"

	"github.com/naver/lobster/pkg/lobster/query"
)

// renamed by log rotate.
type RenamedFileStub struct{}

func (r RenamedFileStub) Keyword() string {
	return "renamedFile"
}
func (r RenamedFileStub) GenerateLogs(conf Config, stopChan chan struct{}) {
	if len(*conf.RenamedFileLogPath) == 0 {
		return
	}

	logFile, err := os.OpenFile(*conf.RenamedFileLogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	var (
		logger    = log.New(logFile, "", 0)
		ticker    = time.NewTicker(100 * time.Millisecond)
		lastInode = getFileInode(*conf.RenamedFileLogPath)
	)

	go func() {
		for {
			select {
			case <-ticker.C:
				currentInode := getFileInode(*conf.RenamedFileLogPath)
				if currentInode != lastInode {
					newLogFile, err := os.OpenFile(*conf.RenamedFileLogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
					if err != nil {
						panic(err)
					}
					logger.SetOutput(newLogFile)
					logFile.Close()
					logFile = newLogFile
					lastInode = currentInode
				}
			case <-stopChan:
				ticker.Stop()
			}
		}
	}()

	generateLogs(logger, conf, stopChan, func(log string) string {
		return fmt.Sprintf("%s %s%s", time.Now().Format(time.RFC3339Nano), r.Keyword(), log)
	})
}

func (r RenamedFileStub) Query() query.Request {
	namespace := os.Getenv("NAMESPACE")
	pod := os.Getenv("POD")

	return query.Request{
		ID:                "loggen-test-renamed",
		Namespaces:        []string{namespace},
		Pods:              []string{pod},
		Page:              -1,
		FilterIncludeExpr: r.Keyword(),
	}
}

func getFileInode(filePath string) uint64 {
	stat, _ := os.Stat(filePath)
	if stat == nil {
		return 0
	}

	return stat.Sys().(*syscall.Stat_t).Ino
}
