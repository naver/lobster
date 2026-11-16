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

package authtoken

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, content string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return file
}

func TestSetHeader(t *testing.T) {
	header := http.Header{}
	if err := SetHeader(header, writeFile(t, "token-a\n")); err != nil {
		t.Fatal(err)
	}
	if got, want := header.Get("Authorization"), "Bearer token-a"; got != want {
		t.Errorf("Authorization = %q, want %q", got, want)
	}

	untouched := http.Header{}
	if err := SetHeader(untouched, ""); err != nil || untouched.Get("Authorization") != "" {
		t.Errorf("SetHeader without a file = %v, %q; want nil and no header", err, untouched.Get("Authorization"))
	}

	if err := SetHeader(http.Header{}, filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("SetHeader accepted a token file that does not exist")
	}
}

func TestLoadTokens(t *testing.T) {
	tokens, err := LoadTokens(writeFile(t, "token-a token-b\n\ntoken-c\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 3 {
		t.Fatalf("LoadTokens() loaded %d tokens, want 3 separated by a space and newlines", len(tokens))
	}

	if _, err := LoadTokens(writeFile(t, "\n")); err == nil {
		t.Error("LoadTokens() accepted a file without tokens")
	}
	if _, err := LoadTokens(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("LoadTokens() accepted a file that does not exist")
	}
}

func TestRequire(t *testing.T) {
	tokens, err := LoadTokens(writeFile(t, "token-a token-b"))
	if err != nil {
		t.Fatal(err)
	}

	handler := Require(tokens, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	tests := []struct {
		name          string
		authorization string
		want          int
	}{
		{"no token", "", http.StatusUnauthorized},
		{"another scheme", "Basic dG9rZW4tYQ==", http.StatusUnauthorized},
		{"token without the scheme", "token-a", http.StatusUnauthorized},
		{"empty token", "Bearer ", http.StatusUnauthorized},
		{"unlisted token", "Bearer token-c", http.StatusUnauthorized},
		{"listed token", "Bearer token-a", http.StatusNoContent},
		{"another listed token", "Bearer token-b", http.StatusNoContent},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/sync", nil)
			if tc.authorization != "" {
				r.Header.Set("Authorization", tc.authorization)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, r)

			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
			if tc.want == http.StatusUnauthorized && rec.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Errorf("WWW-Authenticate = %q, want Bearer", rec.Header().Get("WWW-Authenticate"))
			}
		})
	}
}
