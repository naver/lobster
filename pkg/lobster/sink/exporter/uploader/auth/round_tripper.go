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
	"bytes"
	"io"
	"net/http"

	v1 "github.com/naver/lobster/pkg/operator/api/v1"
	"golang.org/x/oauth2/clientcredentials"
)

type AuthRoundTripper struct {
	Headers   map[string]string
	Host      string
	Transport http.RoundTripper
}

func NewAuthRoundTripper(authType v1.OAuthType, conf *clientcredentials.Config) (http.RoundTripper, error) {
	switch authType {
	case v1.OAuthTypeAuthenzPrincipal:
		return NewAuthenzPrincipalRoundTripper(conf.TokenURL, conf.ClientSecret)
	case v1.OAuthTypeUnencodedCredential:
		return NewUnencodedCredentialRoundTripper(conf.TokenURL, conf.ClientID, conf.ClientSecret)
	default:
		return http.DefaultTransport, nil
	}
}

func (art *AuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Host == art.Host {
		for k, v := range art.Headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := art.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if resp.Body != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body = io.NopCloser(bytes.NewBuffer(data))
	}

	return resp, nil
}
