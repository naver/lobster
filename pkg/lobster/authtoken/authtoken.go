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
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// SetHeader reads the file on every call, so a replaced Secret takes effect
// without a restart.
func SetHeader(header http.Header, file string) error {
	if file == "" {
		return nil
	}

	token, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read auth token: %w", err)
	}
	header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))

	return nil
}

// Tokens holds the accepted tokens by their SHA-256 hashes, so that looking a
// token up takes the same time whatever it shares with an accepted one.
type Tokens map[[sha256.Size]byte]struct{}

// LoadTokens allows several tokens so that one can be replaced without
// refusing syncers.
func LoadTokens(file string) (Tokens, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read auth tokens: %w", err)
	}

	tokens := Tokens{}
	for token := range strings.FieldsSeq(string(data)) {
		tokens[sha256.Sum256([]byte(token))] = struct{}{}
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("%s lists no auth tokens", file)
	}

	return tokens, nil
}

func Require(tokens Tokens, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if _, listed := tokens[sha256.Sum256([]byte(token))]; !ok || !listed {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
