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

package exporter

import (
	"flag"
	"time"
)

type config struct {
	InspectInterval       *time.Duration
	DataPath              *string
	Burst                 *int64
	MaxLookback           *time.Duration
	MinGrpcConnectTimeout *time.Duration
	StoreGrpcServerAddr   *string
	StoreGrpcServerPort   *string
	GrpcMaxCallMsgSize    *int
}

func setup() config {
	inspectInterval := flag.Duration("sink.exporter.inspectInterval", time.Minute, "Log file inspection & sync interval")
	dataPath := flag.String("sink.exporter.dataPath", "/var/lobster", "Database path to store receipts")
	burst := flag.Int64("sink.exporter.burst", 1000000, "Provide lines in and out of busrt per page")
	maxLookback := flag.Duration("sink.exporter.maxLookback", time.Hour, "Limits the time when getting old receipts")
	minGrpcConnectTimeout := flag.Duration("sink.exporter.minGrpcConnectTimeout", time.Second, "Minimum timeout for retrying connection to the store")
	storeGrpcServerAddr := flag.String("sink.exporter.storeGrpcServerAddress", ":11130", "grpc server address in the store")
	grpcMaxCallMsgSize := flag.Int("sink.exporter.grpcMaxCallMsgSize", 10*1024*1024, "The maximum message size (in bytes) allowed for gRPC calls")

	return config{
		InspectInterval:       inspectInterval,
		DataPath:              dataPath,
		Burst:                 burst,
		MaxLookback:           maxLookback,
		MinGrpcConnectTimeout: minGrpcConnectTimeout,
		StoreGrpcServerAddr:   storeGrpcServerAddr,
		GrpcMaxCallMsgSize:    grpcMaxCallMsgSize,
	}
}
