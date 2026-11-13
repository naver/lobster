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

package syncer

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setFlag(t *testing.T, flag *string, value string) {
	t.Helper()
	old := *flag
	*flag = value
	t.Cleanup(func() { *flag = old })
}

func TestRequestSinksSendsTheOperatorToken(t *testing.T) {
	var got []string
	operator := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer operator.Close()
	setFlag(t, conf.LobsterSinkOperator, strings.TrimPrefix(operator.URL, "http://"))

	token := "token-a"
	file := filepath.Join(t.TempDir(), "token")
	if err := os.WriteFile(file, []byte(token), 0o600); err != nil {
		t.Fatal(err)
	}

	setFlag(t, conf.AuthTokenFile, file)
	if _, err := NewSyncer().requestSinks(); err != nil {
		t.Fatal(err)
	}

	setFlag(t, conf.AuthTokenFile, filepath.Join(t.TempDir(), "missing"))
	if _, err := NewSyncer().requestSinks(); err == nil {
		t.Error("requestSinks() with a missing token file returned no error")
	}

	if want := "Bearer " + token; len(got) != 1 || got[0] != want {
		t.Errorf("Authorization headers = %q, want [%q] (no request with a missing token file)", got, want)
	}
}
