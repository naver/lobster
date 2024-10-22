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

package store

import (
	"flag"
	"time"
)

type config struct {
	RetentionSize           *int64
	RetentionTime           *time.Duration
	StoreRootPath           *string
	BlockSize               *int64
	SoftLimitRatioForDisk   *float64
	SoftLimitRatioForBlocks *float64
	ReqMaxBurst             *int64
	ReqCooldownDuration     *time.Duration
	PageBurst               *int
	LeakyBucketInterval     *time.Duration
}

func setup() config {
	retentionSize := flag.Int64("store.retentionSize", (1 << 31), "Max retention size per container logs")
	retentionTime := flag.Duration("store.retentionTime", 7*24*time.Hour, "Max retention time to keep logs")
	storeRootPath := flag.String("store.storeRootPath", "/var/lobster/log", "Path to read/write blocks")
	blockSize := flag.Int64("store.blockSize", (1 << 20), "Block size")
	softLimitRatioForDisk := flag.Float64("store.softLimitForDisk", 0.5, "Size limit of log files")
	softLimitRatioForBlocks := flag.Float64("store.softLimitRatioForBlocks", 0.9, "Ratio of reduction of blocks")
	reqMaxBurst := flag.Int64("store.request.maxBurst", 100000, "The maximum number of requests received by the querier")
	reqCooldownDuration := flag.Duration("store.request.cooldowDuration", 100*time.Millisecond, "Requests that reach the max burst are included in the limiter's count by the cooldown time.")
	pageBurst := flag.Int("store.pageBurst", 1000, "Provide lines in and out of busrt per page")
	leakyBucketInterval := flag.Duration("store.leakyBucketInterval", time.Second, "Interval of flusing logs")

	return config{
		RetentionSize:           retentionSize,
		RetentionTime:           retentionTime,
		StoreRootPath:           storeRootPath,
		BlockSize:               blockSize,
		SoftLimitRatioForDisk:   softLimitRatioForDisk,
		SoftLimitRatioForBlocks: softLimitRatioForBlocks,
		ReqMaxBurst:             reqMaxBurst,
		ReqCooldownDuration:     reqCooldownDuration,
		PageBurst:               pageBurst,
		LeakyBucketInterval:     leakyBucketInterval,
	}
}
