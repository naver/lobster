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

package push

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/metrics"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/store"
	"github.com/naver/lobster/pkg/lobster/util"
)

const (
	Scheme   = "http"
	PathPush = "/push"
)

var (
	headerRealIp = http.CanonicalHeaderKey("X-Real-IP")
	conf         config
	httpClient   = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			IdleConnTimeout:     10 * time.Second,
			MaxIdleConns:        100,
			MaxConnsPerHost:     100,
			MaxIdleConnsPerHost: 100,
			DisableKeepAlives:   false,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 60 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 60 * time.Second,
			WriteBufferSize:     (1 << 20),
		},
	}
)

func init() {
	conf = setup()
	log.Println("push configuration is loaded")
}

func Run(store *store.Store, localAddr string, stopChan chan struct{}) {
	glog.V(3).Infof("local address : %s", localAddr)
	go func() {
		ticker := time.NewTicker(*conf.PushInterval)

		defer func() {
			ticker.Stop()
		}()

		for {
			select {
			case <-ticker.C:
				endpoints, err := util.LookupEndpoints(*conf.LobsterQueryService)
				if err != nil {
					glog.Error(err)
					continue
				}

				if len(endpoints) == 0 {
					glog.Info("no endpoints are found")
					continue
				}

				pushChunks(localAddr, endpoints, store.GetChunks())

			case <-stopChan:
				glog.Info("stop pushing")
				return
			}
		}
	}()
}

func pushChunks(localAddr string, endpoints []string, chunksToPush []model.Chunk) {
	lastIndex := len(chunksToPush) - 1
	chunks := []model.Chunk{}

	for i, chunk := range chunksToPush {
		chunks = append(chunks, chunk)

		if *conf.MaxChunksToPush <= len(chunks) || i == lastIndex {
			data, err := json.Marshal(chunks)
			if err != nil {
				glog.Error(err)
				continue
			}
			if err := push(localAddr, endpoints, data); err != nil {
				glog.Error(err)
				metrics.AddPushError()
				return
			}
			glog.V(3).Infof("push %d chunks %d bytes\n", len(chunks), len(data))
			chunks = chunks[:0]
		}
	}
}

func push(localAddr string, endpoints []string, data []byte) error {
	for _, endpoint := range endpoints {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s%s", endpoint, PathPush), io.NopCloser(bytes.NewBuffer(data)))
		if err != nil {
			return err
		}

		req.Header.Add(headerRealIp, localAddr)

		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		if resp != nil {
			_, err := io.Copy(io.Discard, resp.Body)
			if err != nil {
				return err
			}
			_ = resp.Body.Close()
		}
	}

	return nil
}
