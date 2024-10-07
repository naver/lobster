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

package log

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/query"
	"github.com/naver/lobster/pkg/lobster/server/errors"
)

const (
	PathLogSeries = "/logs/series"
)

type SeriesHandler struct {
	Querier query.Queryable
}

// ServeHTTP
//
//	@Summary		Get series within range
//	@Description	Get series for conditions
//	@Tags			Post
//	@Produce		json
//	@Param			version	path		string			true	"v1 or v2"
//	@Param			request	body		query.Request	true	"request parameters"
//	@Success		200		{object}	query.Response
//	@Success		204		{string}	string	"No chunks"
//	@Failure		400		{string}	string	"Invalid parameters"
//	@Failure		405		{string}	string	"Method not allowed"
//	@Failure		429		{string}	string	"too many requests"
//	@Failure		500		{string}	string	"Failed to read logs"
//	@Router			/api/{version}/logs/series [post]
func (h SeriesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req, err := parseRequest(r)
	if err != nil {
		glog.Info(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Querier.Validate(req); err != nil {
		glog.Info(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	numOfChunk, seriesData, err := h.Querier.GetSeriesInBlocksWithinRange(req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	if numOfChunk == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	data, err := json.Marshal(query.Response{
		SeriesData: &seriesData,
	})
	if err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		glog.Error(err)
	}
}
