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
	"errors"
	"log"
	"net"
	"strings"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/metrics"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/querier"
	"github.com/naver/lobster/pkg/lobster/querier/broker"
	"github.com/naver/lobster/pkg/lobster/query"

	logHandler "github.com/naver/lobster/pkg/lobster/server/handler/log"
)

var conf config

type Querier struct {
	broker.Broker
}

func init() {
	conf = setup()
	log.Println("global querier configuration is loaded")
}

func NewQuerier() *Querier {
	remoteAddrs := []broker.RemoteAddr{}
	clusters := []string{}

	for _, info := range *conf.LobsterQueries {
		part := strings.Split(info, "|")
		splited := strings.Split(part[1], ":")

		if _, err := net.LookupHost(splited[0]); err != nil {
			glog.Infof("skip host %s.", part[1])
			continue
		}

		remoteAddrs = append(remoteAddrs, broker.RemoteAddr{
			Cluster: part[0],
			Address: part[1],
		})
		clusters = append(clusters, part[0])
	}

	glog.Infof("actual clusters after host lookup: %s", strings.Join(clusters, ","))
	return &Querier{
		Broker: broker.NewBroker(remoteAddrs),
	}
}

func (q *Querier) GetChunksWithinRange(req query.Request) (chunks []model.Chunk, err error) {
	chunks, err = q.RequestChunksWithinRange(req, true)
	if err != nil {
		return
	}

	glog.V(3).Infof("%d chunks | %s", len(chunks), req.String())
	return
}

func (q *Querier) GetSeriesInBlocksWithinRange(req query.Request) (numOfChunk int, series model.SeriesData, err error) {
	var (
		chunks  []model.Chunk
		results []querier.FetchResult
	)

	chunks, err = q.RequestChunksWithinRange(req, true)
	if err != nil {
		return
	}

	results, err = querier.Fetch(req, chunks, logHandler.PathLogSeries)
	if err != nil {
		return
	}

	series = querier.NewSeriesBuilder(results).
		Merge().
		Build()

	numOfChunk = len(chunks)

	glog.V(3).Infof("chunks %d | fetched %d | series %d | %s", numOfChunk, len(results), len(series), req.String())
	return
}

func (q *Querier) GetBlocksWithinRange(req query.Request) (data []byte, numOfChunk int, pageInfo model.PageInfo, err error) {
	var (
		chunks           []model.Chunk
		results          []querier.FetchResult
		isPartialEntries bool
		limit            = *conf.ContentsLimit
	)

	if req.ContentsLimit > 0 {
		limit = req.ContentsLimit
	}

	chunks, err = q.RequestChunksWithinRange(req, true)
	if err != nil {
		return
	}

	results, pageInfo, err = querier.FetchLogEntries(req, chunks, limit)
	if err != nil {
		return
	}

	data, isPartialEntries = querier.NewEntryBuilder(results, limit).
		Merge(querier.ParseEntryRaw).
		SortAscending().
		BuildRawLogs()

	if isPartialEntries || pageInfo.IsPartialContents {
		pageInfo.IsPartialContents = true
		metrics.IncreasePartialResponseCount()
	}

	numOfChunk = len(chunks)

	glog.V(3).Infof("chunks %d | fetched %d | data %d | %s", numOfChunk, len(results), len(data), req.String())
	return
}

func (q *Querier) GetEntriesWithinRange(req query.Request) (data []model.Entry, numOfChunk int, pageInfo model.PageInfo, err error) {
	var (
		chunks           []model.Chunk
		results          []querier.FetchResult
		isPartialEntries bool
		limit            = *conf.ContentsLimit
	)

	if req.ContentsLimit > 0 {
		limit = req.ContentsLimit
	}

	chunks, err = q.RequestChunksWithinRange(req, true)
	if err != nil {
		return
	}

	results, pageInfo, err = querier.FetchLogEntries(req, chunks, limit)
	if err != nil {
		return
	}

	data, isPartialEntries = querier.NewEntryBuilder(results, limit).
		Merge(querier.ParseEntry).
		SortAscending().
		Build()

	if isPartialEntries || pageInfo.IsPartialContents {
		pageInfo.IsPartialContents = true
		metrics.IncreasePartialResponseCount()
	}

	numOfChunk = len(chunks)

	glog.V(3).Infof("chunks %d | fetched %d | data %d | %s", numOfChunk, len(results), len(data), req.String())
	return
}

func (q Querier) Validate(req query.Request) error {
	if !req.HasNamespace() && !req.HasNamespaces() {
		return errors.New("invalid namespace")
	}

	return nil
}
