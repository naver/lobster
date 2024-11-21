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

package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	v1 "github.com/naver/lobster/pkg/operator/server/api/v1"
	"github.com/naver/lobster/pkg/operator/server/controller"
)

const (
	Scheme  = "http"
	PathApi = "/api/v1"
)

const (
	PathSinks            = "/namespaces/{namespace}/sinks"
	PathSpecificSink     = "/namespaces/{namespace}/sinks/{name}"
	PathSinkContentsRule = "/namespaces/{namespace}/sinks/{name}/rules/{rule}"
)

type sinkParam struct {
	Namespace string
	Name      string
	Type      string
	Rule      string
}

type SinkHandler struct {
	Ctrl   controller.SinkController
	Logger logr.Logger
}

func (h SinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func(r *http.Request) {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			h.Logger.Error(err, "failed to discard body")
		}
		if err := r.Body.Close(); err != nil {
			h.Logger.Error(err, "failed to close body")
		}
	}(r)

	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPut:
		h.handlePut(w, r)
	case http.MethodDelete:
		h.handleDelete(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGet
//
//	@Summary	List sinks
//	@Tags		Get
//	@Produce	json
//	@Param		namespace	path		string	true	"namespace name"
//	@Param		name		path		string	true	"sink name"
//	@Success	200			{object}	[]v1.Sink
//	@Failure	400			{string}	string	"Invalid parameters"
//	@Failure	405			{string}	string	"Method not allowed"
//	@Failure	500			{string}	string	"Failed to get sink"
//	@Router		/api/v1/namespaces/{namespace}/sinks/{name} [get]
func (h SinkHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	p, err := parseParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sinks, err := h.Ctrl.List(p.Namespace, p.Name, p.Type)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(sinks)
	if err != nil {
		handleError(w, err)
		return
	}

	if _, err := w.Write(data); err != nil {
		h.Logger.Error(err, "failed to write contents")
	}
}

// handlePut
//
//	@Summary	Put log sink
//	@Tags		Put
//	@Accept		json
//	@Param		namespace	path		string	true	"namespace name"
//	@Param		name		path		string	true	"sink name"
//	@Param		sink		body		v1.Sink	true	"sink contentd; Each content in array must be unique"
//	@Success	200			{string}	string	""
//	@Success	201			{string}	string	"Created successfully"
//	@Failure	400			{string}	string	"Invalid parameters"
//	@Failure	422			{string}	string	"Restricted by limits"
//	@Failure	405			{string}	string	"Method not allowed"
//	@Failure	500			{string}	string	"Failed to get sink content"
//	@Router		/api/v1/namespaces/{namespace}/sinks/{name} [put]
func (h SinkHandler) handlePut(w http.ResponseWriter, r *http.Request) {
	p, err := parseParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(p.Name) == 0 {
		http.Error(w, "should set `name`", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sink := v1.Sink{
		Namespace: p.Namespace,
		Name:      p.Name,
	}
	if err := json.Unmarshal(data, &sink); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := sink.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	isCreated, err := h.Ctrl.Put(sink)
	if err != nil {
		handleError(w, err)
		return
	}
	if isCreated {
		w.WriteHeader(http.StatusCreated)
	}
}

// handleDeleteRule
//
//	@Summary	Delete sink
//	@Tags		Delete
//	@Param		namespace	path		string	true	"namespace name"
//	@Param		name		path		string	true	"sink name"
//	@Param		rule		path		string	true	"log export rule name to delete"
//	@Success	200			{string}	string	""
//	@Failure	400			{string}	string	"Invalid parameters"
//	@Failure	404			{string}	string	"Not found"
//	@Failure	405			{string}	string	"Method not allowed"
//	@Failure	500			{string}	string	"Failed to delete sink"
//	@Router		/api/v1/namespaces/{namespace}/sinks/{name}/rules/{rule} [delete]
func (h SinkHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	p, err := parseParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(p.Name) == 0 {
		http.Error(w, "should set `name`", http.StatusBadRequest)
		return
	}

	if len(p.Type) == 0 && len(p.Rule) == 0 {
		if err := h.Ctrl.Delete(p.Namespace, p.Name); err != nil {
			handleError(w, err)
		}
		return
	}

	if err := h.Ctrl.DeleteContent(p.Namespace, p.Name, p.Rule); err != nil {
		handleError(w, err)
	}
}

func parseParam(r *http.Request) (sinkParam, error) {
	vars := mux.Vars(r)
	p := sinkParam{
		Namespace: vars["namespace"],
		Name:      vars["name"],
		Rule:      vars["rule"],
		Type:      r.URL.Query().Get("type"),
	}

	if len(p.Namespace) == 0 {
		return p, fmt.Errorf("should set `namespace`")
	}

	return p, nil
}

func handleError(w http.ResponseWriter, err error) {
	switch err {
	case controller.ErrImproperParam:
		fallthrough
	case controller.ErrUnsupportedType:
		http.Error(w, err.Error(), http.StatusBadRequest)
	case controller.ErrNotFound:
		http.Error(w, err.Error(), http.StatusNotFound)
	case controller.ErrUnprocessableEntity:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
