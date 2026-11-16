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

package server

import (
	"context"
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/naver/lobster/pkg/lobster/authtoken"
	"github.com/naver/lobster/pkg/operator/server/handler"
)

// noSinks answers List with no LobsterSinks, the only client call /sync makes.
type noSinks struct {
	client.Client
}

func (noSinks) List(context.Context, client.ObjectList, ...client.ListOption) error {
	return nil
}

func TestAuthToken(t *testing.T) {
	token := "token-a"
	tokens := authtoken.Tokens{sha256.Sum256([]byte(token)): {}}

	tests := []struct {
		name   string
		tokens authtoken.Tokens
		path   string
		token  string
		want   int
	}{
		{"without tokens /sync needs no token", nil, handler.PathSync, "", http.StatusNoContent},
		{"with tokens /sync refuses a request without a token", tokens, handler.PathSync, "", http.StatusUnauthorized},
		{"with tokens /sync serves a listed token", tokens, handler.PathSync, token, http.StatusNoContent},
		{"with tokens /health needs no token", tokens, "/health", "", http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.token != "" {
				r.Header.Set("Authorization", "Bearer "+tc.token)
			}
			rec := httptest.NewRecorder()
			setupRouter(noSinks{}, logr.Discard(), tc.tokens).ServeHTTP(rec, r)

			if rec.Code != tc.want {
				t.Fatalf("GET %s = %d, want %d", tc.path, rec.Code, tc.want)
			}
		})
	}
}
