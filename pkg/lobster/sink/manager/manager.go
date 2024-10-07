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

package manager

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/client"
	"github.com/naver/lobster/pkg/lobster/metrics"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/sink/indexer"
	"github.com/naver/lobster/pkg/lobster/sink/order"
	orderSync "github.com/naver/lobster/pkg/lobster/sink/sync"
	"github.com/naver/lobster/pkg/lobster/util"
)

var (
	conf config
)

func init() {
	conf = setup()
	log.Println("sink request configuration is loaded")
}

type SinkManager struct {
	client   client.Client
	sinkType string
	cache    sync.Map
}

func NewSinkManager(sinkType string) SinkManager {
	client, err := client.New()
	if err != nil {
		panic(err)
	}
	return SinkManager{
		client:   client,
		sinkType: sinkType,
		cache:    sync.Map{},
	}
}

func (m *SinkManager) Load(key string) ([]order.Order, bool) {
	v, ok := m.cache.Load(key)
	if !ok {
		return nil, false
	}

	return v.([]order.Order), true
}

func (m *SinkManager) Range(receiver func(string, order.Order)) {
	m.cache.Range(func(k, v interface{}) bool {
		for _, order := range v.([]order.Order) {
			receiver(k.(string), order)
		}
		return true
	})
}

func (m *SinkManager) Update(chunks []model.Chunk, start, end time.Time) error {
	var (
		tStart    = util.Timestamp{Time: start}
		tEnd      = util.Timestamp{Time: end}
		preorders []order.Order
	)

	indexer := indexer.New(chunks)

	receivedPreOrders, err := orderSync.Request(*conf.SyncerAddress, m.sinkType, indexer.GetNamespaces())
	if err != nil {
		metrics.AddSinkRequestFailureCount()
		return err
	}
	preorders = append(preorders, receivedPreOrders...)

	glog.V(3).Infof("Requests preorders for [%s] | got %d preorders", strings.Join(indexer.GetNamespaces(), ","), len(preorders))
	metrics.SetReceivingPreorders(float64(len(preorders)))

	orders := order.NewOrdersFromPreorders(preorders, indexer, tStart, tEnd)
	glog.V(3).Infof("%+v", orders)

	m.update(orders)
	metrics.SetOrders(len(orders))

	return nil
}

func (m *SinkManager) update(orders map[string][]order.Order) {
	for k, v := range orders {
		m.cache.Store(k, v)
	}

	m.cache.Range(func(key, value interface{}) bool {
		if _, ok := orders[key.(string)]; !ok {
			m.cache.Delete(key.(string))
		}
		return true
	})
}
