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

package sync

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/naver/lobster/pkg/lobster/syncer"
)

const (
	Scheme   = "http"
	PathSync = "/sync/{type}"
)

type SyncHandler struct {
	Syncer *syncer.Syncer
}

func (h SyncHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sinkType := mux.Vars(r)["type"]
	if err := h.Syncer.Validate(sinkType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	namespaces := []string{}
	if err := json.Unmarshal(data, &namespaces); err != nil {
		http.Error(w, "Failed to unmarshal data", http.StatusBadRequest)
		return
	}

	orders := h.Syncer.GetPreorders(namespaces, sinkType)
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	contents, err := json.Marshal(orders)
	if err != nil {
		http.Error(w, "Failed to read log metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(contents); err != nil {
		glog.Error(err)
	}
}
