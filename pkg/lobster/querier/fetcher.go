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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/query"
	"github.com/naver/lobster/pkg/lobster/server/errors"
	logHandler "github.com/naver/lobster/pkg/lobster/server/handler/log"
)

type FetchResult struct {
	model.Chunk
	response query.Response
	err      error
}

type Fetcher struct {
	client *http.Client
}

func NewFetcher(timeout, responseHeaderTimeout time.Duration) Fetcher {
	return Fetcher{
		&http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				IdleConnTimeout:     10 * time.Second,
				MaxIdleConns:        100,
				MaxConnsPerHost:     100,
				MaxIdleConnsPerHost: 100,
				Dial: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 10 * time.Second,
				}).Dial,
				ResponseHeaderTimeout: responseHeaderTimeout,
				TLSHandshakeTimeout:   10 * time.Second,
				WriteBufferSize:       (1 << 20),
				ReadBufferSize:        (1 << 20),
			},
		},
	}
}

func (f Fetcher) GetLogEntries(req query.Request, chunks []model.Chunk, limit uint64) ([]FetchResult, model.PageInfo, error) {
	results, err := f.Fetch(req, chunks, logHandler.PathLogSeries)
	if err != nil {
		return results, model.PageInfo{}, err
	}

	series := NewSeriesBuilder(results).
		Merge().
		Build()

	burst := req.Burst
	if burst == 0 {
		burst = *conf.PageBurst
	}

	subReq, pageInfo, size, err := query.MakeSubQuery(req, series, int64(burst))
	if err != nil {
		return results, model.PageInfo{}, err
	}

	chunksToFetch := chunks
	if limit < size {
		chunksToFetch, pageInfo.IsPartialContents = limitChunksBySize(req, chunks, series, limit)
	}

	results, err = f.Fetch(subReq, chunksToFetch, logHandler.PathLogRange)
	if err != nil {
		return results, model.PageInfo{}, err
	}

	return results, pageInfo, nil
}

func (f Fetcher) Fetch(req query.Request, chunks []model.Chunk, urlPath string) ([]FetchResult, error) {
	results := []FetchResult{}
	channel := make(chan FetchResult)

	for _, chunk := range chunks {
		go func(req query.Request, c model.Chunk, channel chan FetchResult) {
			r := req
			r.PodUid = c.PodUid
			r.Container = c.Container
			r.Source = c.Source
			result := FetchResult{c, query.Response{}, nil}

			body, err := json.Marshal(r)
			if err != nil {
				glog.Error(err)
				return
			}

			defer func(channel chan FetchResult, result *FetchResult) {
				channel <- *result
			}(channel, &result)

			resp, err := f.client.Do(&http.Request{
				Method: http.MethodPost,
				URL: &url.URL{
					Scheme: logHandler.Scheme,
					Host:   c.StoreAddr,
					Path:   fmt.Sprintf("/api/%s%s", req.Version, urlPath),
				},
				Body: io.NopCloser(bytes.NewBuffer(body)),
			})
			if err != nil {
				result.err = err
				return
			}
			defer func() { _ = resp.Body.Close() }()

			b, err := io.ReadAll(resp.Body)
			if err != nil {
				result.err = err
				return
			}

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
				result.err = errors.ErrorByStatusCode(resp.StatusCode)
				return
			}

			if err := json.Unmarshal(b, &result.response); err != nil {
				glog.Errorf("%s | %s", err.Error(), string(b))
			}
		}(req, chunk, channel)
	}

	var lastError error

	for i := 0; i < len(chunks); i++ {
		r := <-channel
		if r.err != nil {
			lastError = r.err
			continue
		}
		results = append(results, r)
	}

	return results, lastError
}

func limitChunksBySize(req query.Request, chunks []model.Chunk, seriesData model.SeriesData, limit uint64) ([]model.Chunk, bool) {
	result := []model.Chunk{}
	chunkMap := map[string]model.Chunk{}
	total := uint64(0)

	for _, chunk := range chunks {
		chunkMap[chunk.Key()] = chunk
	}

	for _, series := range seriesData {
		chunk, ok := chunkMap[series.ChunkKey]
		if !ok {
			continue
		}
		result = append(result, chunk)
		total = total + series.SizeWithinRange(req.Start.Time, req.End.Time)
		if total > limit {
			return result, true
		}
	}

	return result, false
}
