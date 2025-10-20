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

package syncer

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/query"
	"github.com/naver/lobster/pkg/lobster/sink/order"
	sinkV1 "github.com/naver/lobster/pkg/operator/api/v1"
	v1 "github.com/naver/lobster/pkg/operator/server/api/v1"
)

const (
	scheme   = "http"
	pathSync = "/sync"
)

var conf config

func init() {
	conf = setup()
	log.Println("syncer configuration is loaded")
}

type Syncer struct {
	cache sync.Map
}

func NewSyncer() *Syncer {
	return &Syncer{sync.Map{}}
}

func (r *Syncer) Validate(sinkType string) error {
	if sinkType != sinkV1.LogExportRules && sinkType != sinkV1.LogMetricRules {
		return fmt.Errorf("`%s` type is not supported", sinkType)
	}

	return nil
}

func (r *Syncer) GetPreorders(namespaces []string, sinkType string) []order.Order {
	orders := []order.Order{}

	for _, ns := range namespaces {
		cached, ok := r.cache.Load(ns)
		if !ok {
			continue
		}

		for _, order := range cached.([]order.Order) {
			if order.SinkType != sinkType {
				continue
			}
			orders = append(orders, order)
		}
	}

	return orders
}

func (r *Syncer) Run(stopChan chan struct{}) {
	ticker := time.NewTicker(*conf.SyncInterval)

	go func() {
		defer func() {
			ticker.Stop()
		}()
		for {
			select {
			case <-ticker.C:
				if err := r.Sync(); err != nil {
					glog.Error(err)
				}
			case <-stopChan:
				glog.Info("stop syncer")
				return
			}
		}
	}()
}

func (r *Syncer) Sync() error {
	sinks, err := r.requestSinks()
	if err != nil {
		return err
	}
	glog.Infof("got %d sinks", len(sinks))

	preorderMap := r.mapPreordersFromSinks(sinks)

	glog.Infof("got %d preorderMap", len(preorderMap))

	for ns, preorders := range preorderMap {
		r.cache.Store(ns, preorders)
	}

	r.cache.Range(func(ns, value interface{}) bool {
		key := ns.(string)
		if _, ok := preorderMap[key]; !ok {
			r.cache.Delete(key)
		}
		return true
	})

	return nil
}

func (r *Syncer) mapPreordersFromSinks(sinks []v1.Sink) map[string][]order.Order {
	preorderMap := map[string][]order.Order{}

	for _, sink := range sinks {
		rules := sink.ListSinkRules()

		for _, rule := range rules {
			request := query.NewRequestFromFilter(rule.GetFilter())
			targetNamespace := rule.GetNamespace()

			preorderMap[targetNamespace] = append(preorderMap[targetNamespace], order.NewOrder(sink, rule, request))
		}
	}

	return preorderMap
}

func (r *Syncer) requestSinks() ([]v1.Sink, error) {
	data := []v1.Sink{}

	resp, err := http.DefaultClient.Do(&http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: scheme,
			Host:   *conf.LobsterSinkOperator,
			Path:   pathSync,
		},
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	})
	if err != nil {
		return data, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNoContent {
			return data, nil
		}
		return data, fmt.Errorf("invalid status code %d", resp.StatusCode)
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return data, err
	}

	return data, nil
}
