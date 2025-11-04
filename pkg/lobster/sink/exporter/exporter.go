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
	"context"
	"log"
	"time"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/client"
	"github.com/naver/lobster/pkg/lobster/logline"
	"github.com/naver/lobster/pkg/lobster/metrics"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/proto"
	"github.com/naver/lobster/pkg/lobster/query"
	"github.com/naver/lobster/pkg/lobster/sink/exporter/counter"
	"github.com/naver/lobster/pkg/lobster/sink/exporter/uploader"
	"github.com/naver/lobster/pkg/lobster/sink/exporter/uploader/auth"
	"github.com/naver/lobster/pkg/lobster/sink/manager"
	"github.com/naver/lobster/pkg/lobster/sink/order"
	"github.com/naver/lobster/pkg/lobster/store"
	"github.com/naver/lobster/pkg/lobster/util"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	sinkV1 "github.com/naver/lobster/pkg/operator/api/v1"
)

var conf config

func init() {
	conf = setup()
	log.Println("exporter configuration is loaded")
}

type LogExporter struct {
	counter        counter.Counter
	store          *store.Store
	sinkManager    manager.SinkManager
	client         client.Client
	tokenManager   *auth.TokenManager
	grpcClient     proto.ChunkServiceClient
	protoConverter proto.Converter
}

func NewLogExporter(store *store.Store) LogExporter {
	client, err := client.New()
	if err != nil {
		glog.Fatal(err)
	}

	conn, err := grpc.NewClient(
		*conf.StoreGrpcServerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{MinConnectTimeout: time.Second}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(*conf.GrpcMaxCallMsgSize),
			grpc.MaxCallSendMsgSize(*conf.GrpcMaxCallMsgSize),
		),
	)
	if err != nil {
		glog.Fatal(err)
	}

	return LogExporter{
		counter.NewCounter(*conf.DataPath),
		store,
		manager.NewSinkManager(sinkV1.LogExportRules),
		client,
		auth.NewTokenManager(),
		proto.NewChunkServiceClient(conn),
		proto.Converter{},
	}
}

func (e *LogExporter) Run(stopChan chan struct{}) {
	inspectTicker := time.NewTicker(*conf.InspectInterval)

	for {
		select {
		case <-inspectTicker.C:
			now := time.Now()
			current := now.Truncate(time.Second)

			if err := e.initStore(current); err != nil {
				glog.Error(err)
				continue
			}
			if err := e.sinkManager.Update(e.store.GetChunks()); err != nil {
				glog.Error(err)
			}

			e.sinkManager.Range(func(key string, order order.Order) {
				uploader, err := uploader.New(order, e.tokenManager)
				if err != nil {
					glog.Error(err)
					return
				}

				if errList := uploader.Validate(); !errList.IsEmpty() {
					glog.Error(errList.String())
					metrics.AddSinkFailure(order.Request, order.SinkNamespace, order.SinkName, uploader.Type(), uploader.Name())
					return
				}

				chunk, err := e.loadAndStoreChunkIfExist(order.Request.Source, order.Request.PodUid, order.Request.Container)
				if err != nil {
					glog.Error(err)
					return
				}

				exportedBytes, err := e.export(current, uploader, order, *chunk)
				if err != nil {
					glog.Errorf("%s : %v", err.Error(), order.Request)
					metrics.AddSinkFailure(order.Request, order.SinkNamespace, order.SinkName, uploader.Type(), uploader.Name())
				}

				metrics.AddSinkLogBytes(order.Request, order.SinkNamespace, order.SinkName, uploader.Type(), uploader.Name(), float64(exportedBytes))
			})

			metrics.ClearSinkMetrics()
			e.store.Clear()
			e.counter.Clean(current)
			metrics.ObserveExporterHandleSeconds(time.Since(now).Seconds())
		case <-stopChan:
			glog.Info("stop exporter")
			return
		}
	}
}
func (e *LogExporter) initStore(current time.Time) error {
	chunks, err := e.requestChunks(current.Add(-*conf.MaxLookback), current)
	if err != nil {
		return err
	}

	for i := range chunks {
		e.store.StoreChunk(chunks[i].Source, chunks[i].PodUid, chunks[i].Container, &chunks[i])
	}

	return nil
}

func (e *LogExporter) requestChunks(start, end time.Time) ([]model.Chunk, error) {
	resp, err := e.grpcClient.GetChunksWithinRange(context.Background(), &proto.Request{
		Start: timestamppb.New(start),
		End:   timestamppb.New(end),
	})
	if err != nil {
		return nil, err
	}

	return e.protoConverter.ToChunks(resp.ProtoChunk), nil
}

func (e *LogExporter) loadAndStoreChunkIfExist(source model.Source, podUid, container string) (*model.Chunk, error) {
	resp, err := e.grpcClient.GetChunk(context.Background(), &proto.Request{
		Source: &proto.ProtoSource{
			Type: source.Type,
			Path: source.Path,
		},
		PodUid:    podUid,
		Container: container,
	})
	if err != nil {
		return nil, err
	}

	chunks := e.protoConverter.ToChunks(resp.ProtoChunk)
	if len(chunks) == 0 {
		return nil, errors.New("failed to get chunk")
	}

	e.store.StoreChunk(source, podUid, container, &chunks[0])

	return &chunks[0], nil
}

func (e *LogExporter) export(current time.Time, uploader uploader.Uploader, order order.Order, chunk model.Chunk) (int, error) {
	key := order.Key()
	interval := uploader.Interval()
	receipt, ok, err := e.counter.Load(key)
	if err != nil {
		glog.Error(err)
	}
	if !ok {
		receipt = e.counter.Produce(0, current.Add(-interval), interval, current.Add(-interval))
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
	logTs, total, err := e.getAndExportLogs(uploader, order.Request, chunk, start, end)
	if logTs.IsZero() {
		logTs = start
	}

	if total > 0 {
		receipt.Update(total, current, interval, logTs)
	}

	return receipt.ExportBytes, err
}

func (e *LogExporter) isRightTimeToExport(bucketIntervalSeconds float64, exportedBefore, current time.Time) bool {
	return bucketIntervalSeconds <= current.Sub(exportedBefore.Truncate(time.Second)).Seconds()
}

func (e *LogExporter) makeTimeRange(logTime, current time.Time) (time.Time, time.Time) {
	if *conf.MaxLookback < current.Sub(logTime) {
		return current.Add(-*conf.MaxLookback), current
	}

	return logTime.Add(time.Millisecond), current
}

func (e *LogExporter) getAndExportLogs(uploader uploader.Uploader, request query.Request, chunk model.Chunk, start, end time.Time) (time.Time, int, error) {
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

		if err := uploader.Upload(data, chunk, pStart, pEnd); err != nil {
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
		t, err := logline.ParseTimestamp(string(data))
		if err != nil {
			return time.Time{}, errors.Wrap(err, "failed to parse start")
		}

		return t, nil
	}

	return logline.ParseTimestamp(string(data[:index]))
}

func parseEnd(data []byte) (time.Time, error) {
	index := bytes.LastIndexAny(data[:len(data)-2], "\n")
	if index < 0 {
		t, err := logline.ParseTimestamp(string(data))
		if err != nil {
			return time.Time{}, errors.Wrap(err, "failed to parse end")
		}

		return t, nil
	}

	return logline.ParseTimestamp(string(data[index+1:]))
}
