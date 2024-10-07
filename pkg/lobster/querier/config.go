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

package querier

import (
	"flag"
	"time"
)

type config struct {
	StatusCheckInteval  *time.Duration
	ChunkRetentionTime  *time.Duration
	StoreRetentionTime  *time.Duration
	Id                  *int
	Modulus             *uint64
	LookupServicePrefix *string
	PageBurst           *int
	ContentsLimit       *uint64
}

func setup() config {
	statusCheckInteval := flag.Duration("querier.statusCheckInteval", 10*time.Second, "Interval to clean old chunks")
	chunkRetentionTime := flag.Duration("querier.chunkRetentionTime", 7*24*time.Hour, "Max retention time to keep chunks")
	storeRetentionTime := flag.Duration("querier.storeRetentionTime", 30*time.Second, "Max retention time to keep store addresses")
	id := flag.Int("querier.member.id", 0, "ID within modulus range")
	modulus := flag.Uint64("querier.member.modulus", 1, "Value to perform modulo operation on hash result")
	lookupServicePrefix := flag.String("querier.member.lookup-service-prefix", "lobster-query-shard", "Prefix of service of lobster-querier")
	pageBurst := flag.Int("querier.pageBurst", 1000, "Provide lines in and out of busrt per page")
	contentsLimit := flag.Uint64("querier.contentsLimit", 1000*1000*30, "Limit the amount of responsive content per page")

	return config{
		StatusCheckInteval:  statusCheckInteval,
		ChunkRetentionTime:  chunkRetentionTime,
		StoreRetentionTime:  storeRetentionTime,
		Id:                  id,
		Modulus:             modulus,
		LookupServicePrefix: lookupServicePrefix,
		PageBurst:           pageBurst,
		ContentsLimit:       contentsLimit,
	}
}
