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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/naver/lobster/pkg/lobster/sink/order"
)

const (
	scheme   = "http"
	PathSync = "/sync"
)

var (
	client http.Client = http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
)

func Request(syncer, sinkType string, namespaces []string) ([]order.Order, error) {
	result := []order.Order{}

	data, err := json.Marshal(namespaces)
	if err != nil {
		return result, err
	}

	resp, err := client.Do(&http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Scheme: scheme,
			Host:   syncer,
			Path:   fmt.Sprintf("%s/%s", PathSync, sinkType),
		},
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(bytes.NewBuffer(data)),
	})

	if err != nil {
		return result, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNoContent {
			return result, nil
		}
		return result, fmt.Errorf("invalid status code %d", resp.StatusCode)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, err
	}

	return result, nil
}
