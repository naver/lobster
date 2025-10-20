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

package broker

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
	"github.com/naver/lobster/pkg/lobster/server/handler/log"
)

var (
	httpClient = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			IdleConnTimeout:     3 * time.Second,
			MaxIdleConns:        100,
			MaxConnsPerHost:     100,
			MaxIdleConnsPerHost: 100,
			Dial: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 3 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 3 * time.Second,
			WriteBufferSize:     (1 << 20),
			ReadBufferSize:      (1 << 20),
		},
	}
)

type RemoteAddr struct {
	Cluster string
	Address string
}

type Broker struct {
	remoteAddrs []RemoteAddr
}

func NewBroker(remoteAddrs []RemoteAddr) Broker {
	return Broker{remoteAddrs}
}

func (b *Broker) RequestChunksWithinRange(req query.Request, isGlobal bool) ([]model.Chunk, error) {
	results := []model.Chunk{}
	channel := make(chan []model.Chunk)
	expectedChannelLength := len(b.remoteAddrs)
	clusters := map[string]bool{}

	if isGlobal {
		for _, cluster := range req.Clusters {
			clusters[cluster] = true
		}
		if len(clusters) > 0 {
			expectedChannelLength = len(clusters)
		}
	}

	for _, addr := range b.remoteAddrs {
		if _, ok := clusters[addr.Cluster]; len(clusters) > 0 && !ok {
			continue
		}

		go func(addr string) {
			chunks := []model.Chunk{}

			defer func() {
				channel <- chunks
			}()

			body, err := json.Marshal(req)
			if err != nil {
				glog.Error(err)
				return
			}

			resp, err := httpClient.Do(&http.Request{
				Method: http.MethodPost,
				URL: &url.URL{
					Scheme: log.Scheme,
					Host:   addr,
					Path:   fmt.Sprintf("/api/%s%s", req.Version, log.PathLogs),
				},
				Body: io.NopCloser(bytes.NewBuffer(body)),
			})
			if err != nil {
				glog.Error(err)
				return
			}
			if resp != nil {
				defer func() { _ = resp.Body.Close() }()
			}

			if resp.StatusCode != http.StatusOK {
				if resp.StatusCode != http.StatusNoContent {
					glog.Errorf("got %d from %s", resp.StatusCode, addr)
				}
				return
			}

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				glog.Error(err)
				return
			}

			if err := json.Unmarshal(data, &chunks); err != nil {
				glog.Error(err)
				return
			}
		}(addr.Address)
	}

	for i := 0; i < expectedChannelLength; i++ {
		result := <-channel
		results = append(results, result...)
	}

	return results, nil
}
