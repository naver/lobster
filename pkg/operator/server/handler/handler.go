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

	PathSinkContents = "/namespaces/{namespace}/sinks/{name}/{type}" // deprecated
)

type sinkParam struct {
	Namespace  string
	Name       string
	Type       string
	Rule       string
	RuleName   string
	BucketName string // deprecated
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
//	@Summary			List sinks
//	@Tags				Get
//	@Produce			json
//	@Param				namespace	path		string	true	"namespace name"
//	@Param				name		path		string	false	"sink name"
//	@Param				type		path		string	false	"deprecated;sink type (logMetricRules, logExportRules)"
//	@Param				type		query		string	false	"sink type (logMetricRules, logExportRules)"
//	@Success			200			{object}	[]v1.Sink
//	@Failure			400			{string}	string	"Invalid parameters"
//	@Failure			405			{string}	string	"Method not allowed"
//	@Failure			500			{string}	string	"Failed to get sink"
//	@DeprecatedRouter	/api/v1/namespaces/{namespace}/sinks/{name}/{type} [get]
//	@Router				/api/v1/namespaces/{namespace}/sinks/{name} [get]
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
//	@Summary			Put log sink
//	@Tags				Put
//	@Accept				json
//	@Param				namespace	path		string	true	"namespace name"
//	@Param				name		path		string	true	"sink name"
//	@Param				type		path		string	false	"deprecated;sink type (logMetricRules, logExportRules)"
//	@Param				sink		body		v1.Sink	true	"sink contentd; Each content in array must be unique"
//	@Success			200			{string}	string	""
//	@Success			201			{string}	string	"Created successfully"
//	@Failure			400			{string}	string	"Invalid parameters"
//	@Failure			422			{string}	string	"Restricted by limits"
//	@Failure			405			{string}	string	"Method not allowed"
//	@Failure			500			{string}	string	"Failed to get sink content"
//	@DeprecatedRouter	/api/v1/namespaces/{namespace}/sinks/{name}/{type} [put]
//	@Router				/api/v1/namespaces/{namespace}/sinks/{name} [put]
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
		Type:      p.Type, // TODO: remove this
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

// handleDelete
//
//	@Summary			Delete sink
//	@Tags				Delete
//	@Param				namespace	path		string	true	"namespace name"
//	@Param				name		path		string	true	"sink name"
//	@Param				type		path		string	false	"deprecated;sink type (logMetricRules, logExportRules)"
//	@Param				rule		path		string	false	"log export rule name to delete"
//	@Param				ruleName	query		string	false	"deprecated;metric rule name to delete"
//	@Param				bucketName	query		string	false	"deprecated;bucket name to delete"
//	@Success			200			{string}	string	""
//	@Failure			400			{string}	string	"Invalid parameters"
//	@Failure			404			{string}	string	"Not found"
//	@Failure			405			{string}	string	"Method not allowed"
//	@Failure			500			{string}	string	"Failed to delete sink"
//	@DeprecatedRouter	/api/v1/namespaces/{namespace}/sinks/{name}/{type} [delete]
//	@Router				/api/v1/namespaces/{namespace}/sinks/{name}/rules/{rule} [delete]
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

	if err := h.Ctrl.DeleteContent(p.Namespace, p.Name, p.Rule, p.BucketName); err != nil {
		handleError(w, err)
	}
}

func parseParam(r *http.Request) (sinkParam, error) {
	vars := mux.Vars(r)
	p := sinkParam{
		Namespace:  vars["namespace"],
		Name:       vars["name"],
		Type:       vars["type"], // TODO: UI에서 type 제거 후 path 대신 query param을 쓰도록 변경(get에서만 사용)
		Rule:       vars["rule"],
		RuleName:   r.URL.Query().Get("ruleName"),   // TODO: /rule path 사용 후 제거
		BucketName: r.URL.Query().Get("bucketName"), // TODO: /rule path 사용 후 제거
	}

	// TODO: UI에서 query param 대신 path를 쓰도록 변경 후 제거
	if len(p.Rule) == 0 {
		p.Rule = p.RuleName
	}

	// TODO: UI에서 type 제거 후 path 대신 query param을 쓰도록 변경 후 제거
	if len(p.Type) == 0 {
		p.Type = r.URL.Query().Get("type")
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
