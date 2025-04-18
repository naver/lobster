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
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/metrics"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/querier/broker"
	"github.com/naver/lobster/pkg/lobster/query"
	"github.com/naver/lobster/pkg/lobster/util"

	logHandler "github.com/naver/lobster/pkg/lobster/server/handler/log"
)

var conf config

type pushedData struct {
	data       []byte
	sourceAddr string
}

type Querier struct {
	Id       uint64
	Modulus  uint64
	db       Database
	storeMap sync.Map
	buffer   chan pushedData
	broker.Broker
	Fetcher
}

func init() {
	conf = setup()
	log.Println("querier configuration is loaded")
}

func NewQuerier() *Querier {
	db, err := NewDatabase()
	if err != nil {
		panic(err)
	}

	addrs, err := findRemoteQuerierAddrs(*conf.Id)
	if err != nil {
		panic(fmt.Errorf("failed to find services"))
	}

	return &Querier{
		Id:       uint64(*conf.Id),
		Modulus:  *conf.Modulus,
		db:       db,
		storeMap: sync.Map{},
		buffer:   make(chan pushedData, 10000),
		Broker:   broker.NewBroker(addrs),
		Fetcher:  NewFetcher(*conf.FetchTimeout, *conf.FetchResponseHeaderTimeout),
	}
}

func (q *Querier) UpdateChunks(chunks []model.Chunk) {
	if err := q.db.insert(chunks); err != nil {
		glog.Error(err)
	}
}

func (q *Querier) UpdateStoreStatus(storeAddr string) {
	q.storeMap.Store(storeAddr, time.Now())
}

func (q *Querier) GetChunksWithinRange(req query.Request) (chunks []model.Chunk, err error) {
	var receivedChunks []model.Chunk

	chunks, err = q.getLocalChunksWithinRange(req)
	if err != nil {
		return
	}

	if !req.Local {
		req.Local = true
		receivedChunks, err = q.RequestChunksWithinRange(req, false)
		if err != nil {
			return
		}
		chunks = append(chunks, receivedChunks...)
	}

	glog.V(3).Infof("%d chunks | %s", len(chunks), req.String())
	return
}

func (q *Querier) GetSeriesInBlocksWithinRange(req query.Request) (numOfChunk int, series model.SeriesData, err error) {
	var (
		chunks       []model.Chunk
		remoteChunks []model.Chunk
		results      []FetchResult
	)

	chunks, err = q.getLocalChunksWithinRange(req)
	if err != nil {
		return
	}

	req.Local = true
	remoteChunks, err = q.RequestChunksWithinRange(req, false)
	if err != nil {
		return
	}

	chunks = append(chunks, remoteChunks...)
	results, err = q.Fetch(req, chunks, logHandler.PathLogSeries)
	if err != nil {
		return
	}

	series = NewSeriesBuilder(results).
		Merge().
		Build()

	numOfChunk = len(chunks)

	glog.V(3).Infof("chunks %d | fetched %d |  series %d | %s", numOfChunk, len(results), len(series), req.String())
	return
}

func (q *Querier) GetBlocksWithinRange(req query.Request) (data []byte, numOfChunk int, pageInfo model.PageInfo, err error) {
	var (
		chunks           []model.Chunk
		remoteChunks     []model.Chunk
		results          []FetchResult
		isPartialEntries bool
		limit            = *conf.ContentsLimit
	)

	if req.ContentsLimit > 0 {
		limit = req.ContentsLimit
	}

	chunks, err = q.getLocalChunksWithinRange(req)
	if err != nil {
		return
	}

	req.Local = true
	remoteChunks, err = q.RequestChunksWithinRange(req, false)
	if err != nil {
		return
	}

	chunks = append(chunks, remoteChunks...)
	results, pageInfo, err = q.GetLogEntries(req, chunks, limit)
	if err != nil {
		return
	}

	data, isPartialEntries = NewEntryBuilder(results, limit).
		Merge(ParseEntryRaw).
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
		remoteChunks     []model.Chunk
		results          []FetchResult
		isPartialEntries bool
		limit            = *conf.ContentsLimit
	)

	if req.ContentsLimit > 0 {
		limit = req.ContentsLimit
	}

	chunks, err = q.getLocalChunksWithinRange(req)
	if err != nil {
		return
	}

	req.Local = true
	remoteChunks, err = q.RequestChunksWithinRange(req, false)
	if err != nil {
		return
	}

	chunks = append(chunks, remoteChunks...)
	results, pageInfo, err = q.GetLogEntries(req, chunks, limit)
	if err != nil {
		return
	}

	data, isPartialEntries = NewEntryBuilder(results, limit).
		Merge(ParseEntry).
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

func (q *Querier) getLocalChunksWithinRange(req query.Request) ([]model.Chunk, error) {
	chunks := []model.Chunk{}
	localChunks := []model.Chunk{}
	chunkMatcher := NewChunkMatcher(req)

	if len(req.Namespaces) == 0 && len(req.Namespace) == 0 {
		return q.db.getAllChunksWithinRange(req.Start.Time, req.End.Time)
	}

	namespaces := append(req.Namespaces, req.Namespace)

	for _, ns := range namespaces {
		storedChunks, err := q.db.getChunksForNamespaceWithinRange(ns, req.Start.Time, req.End.Time)
		if err != nil {
			return chunks, err
		}

		localChunks = append(localChunks, storedChunks...)
	}

	for _, chunk := range localChunks {
		if !chunkMatcher.IsRequestedChunk(chunk) {
			continue
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

func (q *Querier) Validate(req query.Request) error {
	if !req.HasNamespace() && !req.HasNamespaces() {
		return fmt.Errorf("invalid namespace")
	}

	return nil
}

func (q *Querier) HandlePushData(data []byte, sourceAddr string) {
	q.buffer <- pushedData{data, sourceAddr}
}

func (q *Querier) Run(stopChan chan struct{}) {
	go q.receiveChunks(stopChan)
	go q.handleStatus(stopChan)
}

func (q *Querier) receiveChunks(stopChan chan struct{}) {
	for {
		select {
		case pushed, ok := <-q.buffer:
			if !ok {
				return
			}

			var (
				chunks         = []model.Chunk{}
				chunksToUpdate = []model.Chunk{}
			)

			if err := json.Unmarshal(pushed.data, &chunks); err != nil {
				glog.Errorf("%s | %s", err.Error(), string(pushed.data))
				continue
			}

			for _, chunk := range chunks {
				chunkToUpdate := chunk
				chunkToUpdate.StoreAddr = pushed.sourceAddr
				chunksToUpdate = append(chunksToUpdate, chunkToUpdate)
			}

			q.UpdateChunks(chunksToUpdate)
			q.UpdateStoreStatus(pushed.sourceAddr)

			glog.V(3).Infof("commit %d chunks", len(chunks))
		case <-stopChan:
			glog.Info("stop status push receiver")
		}
	}
}

func (q *Querier) handleStatus(stopChan chan struct{}) {
	ticker := time.NewTicker(*conf.StatusCheckInteval)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			q.handleStoreStatus()

			chunks, err := q.db.getChunks()
			if err != nil {
				glog.Error(err)
			}

			q.handleRetention(chunks)
			q.updateMetrics(chunks)
		case <-stopChan:
			glog.Info("stop status inspection")
		}
	}
}

func (q *Querier) handleStoreStatus() {
	now := time.Now()
	q.storeMap.Range(func(key, value interface{}) bool {
		if *conf.StoreRetentionTime < now.Sub(value.(time.Time)) {
			if err := q.db.deleteByAddr(key.(string)); err != nil {
				glog.Error(err)
			}
			glog.Infof("delete chunks by store addr %s", key.(string))
		}
		return true
	})
}

func (q *Querier) handleRetention(chunks []model.Chunk) {
	deadline := time.Now().Add(-*conf.ChunkRetentionTime)

	for _, chunk := range chunks {
		if chunk.UpdatedAt.Before(deadline) {
			if err := q.db.delete(chunk); err != nil {
				glog.Error(err)
			}
			glog.V(3).Infof("deleted chunk : %s_%s_%s", chunk.Namespace, chunk.Pod, chunk.Container)
		}
	}
}

func (q *Querier) updateMetrics(chunks []model.Chunk) {
	metrics.SetStoredChunks(float64(len(chunks)))
}

func findRemoteQuerierAddrs(excludedId int) ([]broker.RemoteAddr, error) {
	addrs := []broker.RemoteAddr{}
	for i := 0; i < int(*conf.Modulus); i++ {
		if i == excludedId {
			continue
		}
		eps, err := util.LookupEndpoints(fmt.Sprintf("%s-%d", *conf.LookupServicePrefix, i))
		if err != nil {
			return addrs, err
		}
		addrs = append(addrs, broker.RemoteAddr{
			Cluster: "local",
			Address: eps[0],
		})
	}
	return addrs, nil
}
