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
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/query"
	"github.com/naver/lobster/pkg/lobster/server/errors"
)

const (
	PathLogRange = "/logs/range"
)

type RangeHandler struct {
	Querier                         query.Queryable
	ShouldResponseStringContentOnly bool
}

// ServeForStringContent
//
//	@Summary		Get logs within range
//	@Description	Get logs for conditions
//	@Tags			Post
//	@Produce		json
//	@Param			request	body		query.Request	true	"request parameters"
//	@Success		200		{object}	query.Response
//	@Success		204		{string}	string	"No chunks"
//	@Failure		400		{string}	string	"Invalid parameters"
//	@Failure		405		{string}	string	"Method not allowed"
//	@Failure		429		{string}	string	"too many requests"
//	@Failure		500		{string}	string	"Failed to read logs"
//	@Failure		501		{string}	string	"Not supported version"
//	@Router			/api/v1/logs/range [post]
func (h RangeHandler) ServeForStringContent(req query.Request, w http.ResponseWriter, r *http.Request) {
	if err := h.Querier.Validate(req); err != nil {
		glog.Info(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Page == 0 || req.Page < query.LastPageNum {
		http.Error(w, "invalid page number", http.StatusBadRequest)
		return
	}

	content, numOfChunk, pageInfo, err := h.Querier.GetBlocksWithinRange(req)
	if err != nil {
		errors.HandleError(w, err)
		glog.Error(err)
		return
	}

	if numOfChunk == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	data, err := json.Marshal(query.Response{
		Contents: string(content),
		PageInfo: &pageInfo,
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

// ServeForEntriesContent
//
//	@Summary		Get logs within range
//	@Description	Get logs for conditions
//	@Tags			Post
//	@Produce		json
//	@Param			request	body		query.Request	true	"request parameters"
//	@Success		200		{object}	query.ResponseEntries
//	@Success		204		{string}	string	"No chunks"
//	@Failure		400		{string}	string	"Invalid parameters"
//	@Failure		405		{string}	string	"Method not allowed"
//	@Failure		429		{string}	string	"too many requests"
//	@Failure		500		{string}	string	"Failed to read logs"
//	@Failure		501		{string}	string	"Not supported version"
//	@Router			/api/v2/logs/range [post]
func (h RangeHandler) ServeForEntriesContent(req query.Request, w http.ResponseWriter, r *http.Request) {
	if err := h.Querier.Validate(req); err != nil {
		glog.Info(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Page == 0 || req.Page < query.LastPageNum {
		http.Error(w, "invalid page number", http.StatusBadRequest)
		return
	}

	entries, numOfChunk, pageInfo, err := h.Querier.GetEntriesWithinRange(req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	if numOfChunk == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	data, err := json.Marshal(query.ResponseEntries{
		Contents: entries,
		PageInfo: &pageInfo,
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

func (h RangeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	switch req.Version {
	case ApiV1:
		req.Source.Type = model.LogTypeStdStream
		h.ServeForStringContent(req, w, r)
	case ApiV2:
		if h.ShouldResponseStringContentOnly {
			h.ServeForStringContent(req, w, r)
		} else {
			h.ServeForEntriesContent(req, w, r)
		}
	default:
		http.Error(w, "not implemented api version", http.StatusNotImplemented)
		return
	}
}
