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

package web

import (
	"net/http"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/query"
	"github.com/naver/lobster/pkg/lobster/server/handler/log"
)

const webContentsLimit = 1000 * 1000 * 3 // 3mb

type WebHandler struct {
	Addresses []string
	Querier   query.Queryable
}

func (h WebHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	page := newPage()
	req, _ := query.ParseRequestWithUri(r.RequestURI)
	req.Version = log.ApiV2
	req.ContentsLimit = webContentsLimit

	if !req.Start.Time.IsZero() && !req.End.Time.IsZero() {
		chunks, err := h.Querier.GetChunksWithinRange(req)
		if err != nil {
			glog.Error(err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		page.fillPanel(chunks)

		if shouldRespondLogs(req) {
			_, seriesData, err := h.Querier.GetSeriesInBlocksWithinRange(req)
			if err != nil {
				glog.Error(err)
			}

			contents, _, pageInfo, err := h.Querier.GetBlocksWithinRange(req)
			if err != nil {
				glog.Error(err)
			}

			if err := page.fillContents(contents, seriesData, pageInfo); err != nil {
				glog.Error(err)
			}
		}
	}

	if err := page.render(w); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
}

func shouldRespondLogs(req query.Request) bool {
	return req.HasLabels() || req.HasSetNames() || req.HasPods() || req.HasContainers()
}
