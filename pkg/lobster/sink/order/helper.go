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

package order

import (
	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/metrics"
	"github.com/naver/lobster/pkg/lobster/querier"
	"github.com/naver/lobster/pkg/lobster/sink/indexer"
	"github.com/naver/lobster/pkg/lobster/util"
)

type Orders map[string][]Order

func (orders Orders) Put(key string, orderList []Order) {
	if v, ok := orders[key]; ok {
		orders[key] = append(v, orderList...)
	} else {
		orders[key] = orderList
	}
}

func (orders Orders) Append(input Orders) {
	for key, orderList := range input {
		orders.Put(key, orderList)
	}
}

func NewOrdersFromPreorders(preorders []Order, indexer indexer.ChunkIndexer, start, end util.Timestamp) Orders {
	orders := Orders{}

	for _, preorder := range preorders {
		preorder.Request.Start = start
		preorder.Request.End = end
		newOrders, err := newOrdersForChunks(preorder, indexer)
		if err != nil {
			glog.Error("Invalid order : %s : %s_%s_%s", err.Error(), preorder.SinkNamespace, preorder.SinkName, preorder.ContentsName)
			metrics.AddInvalidRequestCount(preorder.SinkNamespace, preorder.SinkName, preorder.ContentsName)
			continue
		}

		orders.Append(newOrders)
	}

	return orders
}

func newOrdersForChunks(preorder Order, indexer indexer.ChunkIndexer) (Orders, error) {
	orders := Orders{}
	matcher := querier.NewChunkMatcher(preorder.Request)
	chunks := indexer.GetChunks(preorder.ContentsNamespace)

	for _, chunk := range chunks {
		if !matcher.IsRequestedChunk(chunk) {
			continue
		}
		order, err := preorder.Clone()
		if err != nil {
			return orders, err
		}

		if err := order.Update(chunk); err != nil {
			return orders, err
		}

		orders.Put(chunk.Key(), []Order{order})
	}

	return orders, nil
}
