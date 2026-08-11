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

package global

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang/glog"
)

const (
	PathStatus = "/status"

	probeTimeout     = 3 * time.Second
	errNotRegistered = "not registered at startup: host lookup failed"
)

// probeClient is separate from the clients that carry log requests; a status
// probe must not wait as long as a log query.
var probeClient = &http.Client{Timeout: probeTimeout}

// ClusterStatus is one configured lobsterQuery entry and whether it answers.
type ClusterStatus struct {
	Cluster   string `json:"cluster"`
	Address   string `json:"address"`
	Connected bool   `json:"connected"`
	Error     string `json:"error,omitempty"`
}

// Status is the connectivity of every configured lobsterQuery entry.
type Status struct {
	Total     int             `json:"total"`
	Connected int             `json:"connected"`
	Clusters  []ClusterStatus `json:"clusters"`
}

// Status probes each configured lobster query and reports whether it answers.
func (q *Querier) Status() Status {
	var (
		wg       sync.WaitGroup
		clusters = make([]ClusterStatus, len(q.remotes))
	)

	for i, r := range q.remotes {
		wg.Add(1)

		go func(i int, r remote) {
			defer wg.Done()
			clusters[i] = probe(r)
		}(i, r)
	}

	wg.Wait()

	status := Status{Total: len(clusters), Clusters: clusters}
	for _, cluster := range clusters {
		if cluster.Connected {
			status.Connected++
		}
	}

	return status
}

func probe(r remote) ClusterStatus {
	status := ClusterStatus{Cluster: r.Cluster, Address: r.Address}

	if !r.Registered {
		status.Error = errNotRegistered
		return status
	}

	resp, err := probeClient.Get(fmt.Sprintf("http://%s/health", r.Address))
	if err != nil {
		status.Error = err.Error()
		return status
	}
	defer func() { _ = resp.Body.Close() }()

	status.Connected = true

	return status
}

type StatusHandler struct {
	Querier *Querier
}

func (h StatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := json.Marshal(h.Querier.Status())
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
