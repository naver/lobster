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

package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
)

/*
 * This is an option to add the header without performing URL encoding.
 * Some servers do not perform URL decoding even when the `application/x-www-form-urlencoded` content type is specified in the header.
 * This behavior is non-compliant with the standard as outlined in RFC 6749 Section 2.3.1.
 * https://datatracker.ietf.org/doc/html/rfc6749#section-2.3.1
 */
func NewUnencodedCredentialRoundTripper(tokenURL, clientId, clientSecret string) (*AuthRoundTripper, error) {
	u, err := url.Parse(tokenURL)
	if err != nil {
		return nil, err
	}

	return &AuthRoundTripper{
		Headers:   map[string]string{"Authorization": fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientId, clientSecret))))},
		Host:      u.Host,
		Transport: http.DefaultTransport,
	}, nil
}
