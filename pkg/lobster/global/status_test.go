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
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/naver/lobster/pkg/lobster/querier/broker"
	"github.com/naver/lobster/pkg/lobster/server"
)

// newFakeQuery serves the health that server.Router() gives every component.
func newFakeQuery(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(server.Router())
}

func hostOf(t *testing.T, rawURL string) string {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}

	return parsed.Host
}

func newRemote(cluster, address string, registered bool) remote {
	return remote{
		RemoteAddr: broker.RemoteAddr{Cluster: cluster, Address: address},
		Registered: registered,
	}
}

func TestStatusConnected(t *testing.T) {
	alive := newFakeQuery(t)
	defer alive.Close()

	querier := &Querier{remotes: []remote{newRemote("alive", hostOf(t, alive.URL), true)}}
	status := querier.Status()

	if status.Total != 1 || status.Connected != 1 {
		t.Fatalf("expected 1/1 connected, got %d/%d", status.Connected, status.Total)
	}
	if !status.Clusters[0].Connected {
		t.Fatalf("expected connected cluster, got %+v", status.Clusters[0])
	}
	if len(status.Clusters[0].Error) > 0 {
		t.Fatalf("expected no error, got %q", status.Clusters[0].Error)
	}
}

func TestStatusUnreachable(t *testing.T) {
	dead := newFakeQuery(t)
	address := hostOf(t, dead.URL)
	dead.Close()

	querier := &Querier{remotes: []remote{newRemote("dead", address, true)}}
	status := querier.Status()

	if status.Connected != 0 {
		t.Fatalf("expected 0 connected, got %d", status.Connected)
	}
	if len(status.Clusters[0].Error) == 0 {
		t.Fatal("expected the request error to be reported")
	}
	if status.Clusters[0].Error == errNotRegistered {
		t.Fatal("a registered host must not be reported as unregistered")
	}
}

// A host that did not resolve at startup is not queried, so it must be reported
// as disconnected without a probe telling otherwise.
func TestStatusNotRegistered(t *testing.T) {
	alive := newFakeQuery(t)
	defer alive.Close()

	querier := &Querier{remotes: []remote{newRemote("ghost", hostOf(t, alive.URL), false)}}
	status := querier.Status()

	if status.Connected != 0 {
		t.Fatalf("expected 0 connected, got %d", status.Connected)
	}
	if status.Clusters[0].Error != errNotRegistered {
		t.Fatalf("expected %q, got %q", errNotRegistered, status.Clusters[0].Error)
	}
}

func TestStatusCounts(t *testing.T) {
	alive := newFakeQuery(t)
	defer alive.Close()

	dead := newFakeQuery(t)
	deadAddress := hostOf(t, dead.URL)
	dead.Close()

	querier := &Querier{remotes: []remote{
		newRemote("alive-1", hostOf(t, alive.URL), true),
		newRemote("alive-2", hostOf(t, alive.URL), true),
		newRemote("dead", deadAddress, true),
		newRemote("ghost", "no-such-host:8080", false),
	}}
	status := querier.Status()

	if status.Total != 4 || status.Connected != 2 {
		t.Fatalf("expected 2/4 connected, got %d/%d", status.Connected, status.Total)
	}

	// The order of the configured entries is kept, so that a caller can match
	// the cluster it asked about.
	for i, expected := range []string{"alive-1", "alive-2", "dead", "ghost"} {
		if status.Clusters[i].Cluster != expected {
			t.Fatalf("expected cluster %q at %d, got %q", expected, i, status.Clusters[i].Cluster)
		}
	}
}

func TestStatusHandler(t *testing.T) {
	alive := newFakeQuery(t)
	defer alive.Close()

	handler := StatusHandler{Querier: &Querier{
		remotes: []remote{newRemote("alive", hostOf(t, alive.URL), true)},
	}}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, PathStatus, nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected a json content type, got %q", contentType)
	}

	status := Status{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if status.Total != 1 || status.Connected != 1 {
		t.Fatalf("expected 1/1 connected, got %d/%d", status.Connected, status.Total)
	}

	// Log contents are served on authenticated routes; this one must not leak
	// anything but addresses and flags.
	for _, field := range []string{"contents", "message", "storeAddr"} {
		if strings.Contains(recorder.Body.String(), field) {
			t.Fatalf("unexpected field %q in %s", field, recorder.Body.String())
		}
	}
}

func TestStatusHandlerRejectsPost(t *testing.T) {
	handler := StatusHandler{Querier: &Querier{}}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, PathStatus, nil))

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", recorder.Code)
	}
}
