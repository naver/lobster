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

package global

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

const MaxBytes = 1024 * 1024 * 1024 // 1 gib

type LobsterQueries []string

func (a *LobsterQueries) String() string {
	return strings.Join(*a, ",")
}

func (a *LobsterQueries) Set(addr string) error {
	*a = append(*a, addr)
	return nil
}

type config struct {
	LobsterQueries             *LobsterQueries
	PageBurst                  *int
	ExportLimit                *int
	ContentsLimit              *uint64
	FetchTimeout               *time.Duration
	FetchResponseHeaderTimeout *time.Duration
}

func setup() config {
	lobsterQueries := &LobsterQueries{}
	flag.Var(lobsterQueries, "global.lobsterQuery", "lobster query address and cluster name separated by '|'; e.g. {cluster}|{address}")
	pageBurst := flag.Int("global.pageBurst", 1000, "Provide lines in and out of busrt per page")
	limit := flag.Int("global.exportlLimit", MaxBytes, fmt.Sprintf("limit in bytes (0 < limit < %d)", MaxBytes))
	contentsLimit := flag.Uint64("global.contentsLimit", 1000*1000*30, "Limit the amount of responsive content per page")
	fetchTimeout := flag.Duration("global.fetchTimeout", 10*time.Second, "Response timeout for log requests")
	fetchResponseHeaderTimeout := flag.Duration("global.fetchResponseHeaderTimeout", 10*time.Second, "Header response timeout for log requests; delays may occur during file reading")

	return config{
		LobsterQueries:             lobsterQueries,
		PageBurst:                  pageBurst,
		ExportLimit:                limit,
		ContentsLimit:              contentsLimit,
		FetchTimeout:               fetchTimeout,
		FetchResponseHeaderTimeout: fetchResponseHeaderTimeout,
	}
}
