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
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/client"
	"github.com/naver/lobster/pkg/lobster/logline"
	"github.com/naver/lobster/pkg/lobster/metrics"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/query"
	"github.com/naver/lobster/pkg/lobster/sink/exporter/bucket"
	"github.com/naver/lobster/pkg/lobster/sink/exporter/counter"
	"github.com/naver/lobster/pkg/lobster/sink/helper"
	"github.com/naver/lobster/pkg/lobster/sink/manager"
	"github.com/naver/lobster/pkg/lobster/sink/order"
	"github.com/naver/lobster/pkg/lobster/store"
	"github.com/naver/lobster/pkg/lobster/util"

	sinkV1 "github.com/naver/lobster/pkg/operator/api/v1"
)

var conf config

func init() {
	conf = setup()
	log.Println("exporter configuration is loaded")
}

type LogExporter struct {
	counter     counter.Counter
	store       *store.Store
	sinkManager manager.SinkManager
	client      client.Client
}

func NewLogExporter(store *store.Store) LogExporter {
	client, err := client.New()
	if err != nil {
		panic(err)
	}
	return LogExporter{
		counter.NewCounter(*conf.DataPath),
		store,
		manager.NewSinkManager(sinkV1.LogExportRules),
		client,
	}
}

func (e *LogExporter) Run(stopChan chan struct{}) {
	inspectTicker := time.NewTicker(*conf.InspectInterval)
	recentOrders := map[string]order.Order{}

	for {
		select {
		case <-inspectTicker.C:
			now := time.Now()
			current := now.Truncate(time.Second)
			newOrders := map[string]order.Order{}

			podMap, err := e.client.GetPods()
			if err != nil {
				glog.Error(err)
				continue
			}

			e.store.InitChunks()
			if err := e.sinkManager.Update(helper.FilterChunksByExistingPods(e.store.GetChunks(), podMap), current.Add(-*conf.InspectInterval), current); err != nil {
				glog.Error(err)
				continue
			}

			e.sinkManager.Range(func(key string, order order.Order) {
				recentOrders[order.Key()] = order
				newOrders[order.Key()] = order
			})

			for _, order := range recentOrders {
				bkt, err := bucket.New(order)
				if err != nil {
					glog.Error(err)
					continue
				}

				if err := bkt.Validate(); err != nil {
					glog.Error(err)
					continue
				}

				chunk := e.store.LoadChunk(order.Request.Source, order.Request.PodUID, order.Request.Container)
				if chunk == nil {
					continue
				}

				exportedBytes, err := e.exportToBucket(current, bkt, order, *chunk)
				if err != nil {
					glog.Errorf("%s : %v", err.Error(), order.Request)
					metrics.AddSinkFailure(order.Request, order.SinkNamespace, order.SinkName, order.SinkType, bkt.Name())
				}

				metrics.AddSinkLogBytes(order.Request, order.SinkNamespace, order.SinkName, order.SinkType, bkt.Name(), float64(exportedBytes))
			}

			recentOrders = copyOrders(newOrders)
			metrics.ClearSinkMetrics()
			e.store.Clear()
			e.counter.Clean()
			metrics.ObserveExporterHandleSeconds(time.Since(now).Seconds())
		case <-stopChan:
			glog.Info("stop exporter")
			return
		}
	}
}

func (e *LogExporter) exportToBucket(current time.Time, bucket bucket.Bucket, order order.Order, chunk model.Chunk) (int, error) {
	key := order.Key()
	interval := bucket.Interval()
	receipt, ok, err := e.counter.Load(key)
	if err != nil {
		glog.Error(err)
	}
	if !ok {
		receipt = e.counter.Produce(0, current.Add(-interval), interval, current)
	}

	defer func(key string, receipt *counter.Receipt) {
		if err := e.counter.Store(key, *receipt); err != nil {
			glog.Error(err)
		}
	}(key, &receipt)

	if !e.isRightTimeToExport(interval.Seconds(), receipt.ExportTime, current) {
		return 0, nil
	}

	start, end := e.makeTimeRange(receipt.LogTime, current)
	logTs, total, err := e.getAndExportLogs(bucket, order.Request, chunk, start, end)

	if logTs.IsZero() {
		logTs = current
	}

	receipt.Update(total, current, interval, logTs)

	return receipt.ExportBytes, err
}

func (e *LogExporter) isRightTimeToExport(bucketIntervalSeconds float64, exportedBefore, current time.Time) bool {
	return bucketIntervalSeconds <= current.Sub(exportedBefore.Truncate(time.Second)).Seconds()
}

func (e *LogExporter) makeTimeRange(logTime, current time.Time) (time.Time, time.Time) {
	if *conf.MaxLookback < current.Sub(logTime) {
		return current.Add(-*conf.MaxLookback), current
	}

	return logTime.Add(time.Nanosecond), current
}

func (e *LogExporter) getAndExportLogs(bucket bucket.Bucket, request query.Request, chunk model.Chunk, start, end time.Time) (time.Time, int, error) {
	ts := time.Time{}
	total := 0
	hasNext := true

	if start.After(end) {
		return ts, total, nil
	}

	request.Start = util.Timestamp{Time: start}
	request.End = util.Timestamp{Time: end}
	request.Page = 1

	_, series, err := e.store.GetSeriesInBlocksWithinRange(request)
	if err != nil {
		return time.Time{}, 0, err
	}

	for hasNext {
		subReq, pageInfo, _, err := query.MakeSubQuery(request, series, *conf.Burst)
		if err != nil {
			return time.Time{}, 0, err
		}

		data, _, _, err := e.store.GetBlocksWithinRange(subReq)
		if err != nil {
			return time.Time{}, 0, err
		}

		if len(data) == 0 {
			return time.Time{}, 0, nil
		}

		pStart, err := parseStart(data)
		if err != nil {
			return time.Time{}, 0, err
		}

		pEnd, err := parseEnd(data)
		if err != nil {
			return time.Time{}, 0, err
		}

		fileName := bucket.FileName(pStart, pEnd)
		dir := bucket.Dir(chunk, start)

		if err := bucket.Flush(data, dir, fileName); err != nil {
			return time.Time{}, 0, err
		}

		request.Page = request.Page + 1
		hasNext = pageInfo.HasNext
		ts = pEnd
		total = total + len(data)
	}

	return ts, total, nil
}

func parseStart(data []byte) (time.Time, error) {
	index := bytes.IndexAny(data, "\n")
	if index < 0 {
		return time.Time{}, fmt.Errorf("failed to parse start")
	}

	return logline.ParseTimestamp(string(data[:index]))
}

func parseEnd(data []byte) (time.Time, error) {
	index := bytes.LastIndexAny(data[:len(data)-2], "\n")
	if index < 0 {
		return time.Time{}, fmt.Errorf("failed to parse end")
	}

	return logline.ParseTimestamp(string(data[index+1:]))
}

func copyOrders(orders map[string]order.Order) map[string]order.Order {
	ret := map[string]order.Order{}

	for k, v := range orders {
		ret[k] = v
	}

	return ret
}
