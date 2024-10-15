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
	"flag"
	"time"
)

type Config struct {
	EnableInspector *bool
	ServerAddr      *string

	LogGenerationInterval *time.Duration
	LogSize               *int
	LogLines              *int
	RenamedFileLogPath    *string

	LobsterQueryEndpoint *string
	Interval             *time.Duration
	WarmUpWait           *time.Duration
}

func NewConfig() Config {
	defer flag.Parse()
	return Config{
		EnableInspector:       flag.Bool("enableInspector", true, "enabled in default"),
		ServerAddr:            flag.String("serverAddr", ":8080", "server address"),
		LogGenerationInterval: flag.Duration("gen.interval", time.Second, "interval to generate logs"),
		LogSize:               flag.Int("gen.size", 1, "size of each log"),
		LogLines:              flag.Int("gen.lines", 1, "number of lines per interval"),
		RenamedFileLogPath:    flag.String("gen.renamedFileLogPath", "", "file path"),
		LobsterQueryEndpoint:  flag.String("inspector.lobsterQueryEndpoint", "lobster-query:8080", "lobster to query"),
		Interval:              flag.Duration("inspector.interval", 10*time.Second, "interval to generate logs"),
		WarmUpWait:            flag.Duration("inspector.warmupWait", 20*time.Second, "wait before running inspector"),
	}
}
