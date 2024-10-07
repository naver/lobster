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

package loggen

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/naver/lobster/pkg/lobster/query"
)

var (
	errConnection    = errors.New("connection_failed")
	errNoContent     = errors.New("no_content")
	errReadBody      = errors.New("read_body_failed")
	errUnmarshalBody = errors.New("unmarshal_body_failed")
)

type client struct {
	*http.Client
}

func newClient() client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.IdleConnTimeout = time.Second
	return client{&http.Client{
		Transport: t,
	}}
}

func (c *client) RequestLogEntries(queryEndpoint string, req query.Request) (query.ResponseEntries, error) {
	logger := log.New(os.Stderr, "", 0)
	queryResp := query.ResponseEntries{}

	body, err := json.Marshal(req)
	if err != nil {
		return queryResp, err
	}

	resp, err := c.Do(&http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Scheme: "http",
			Host:   queryEndpoint,
			Path:   "/api/v2/logs/range",
		},
		Body: io.NopCloser(bytes.NewBuffer(body)),
	})
	if err != nil {
		return queryResp, errConnection
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return queryResp, errReadBody
	}

	if resp.StatusCode != 200 {
		if resp.StatusCode == http.StatusNoContent {
			return queryResp, errNoContent
		}

		if len(data) != 0 {
			logger.Printf("[%d] %s", resp.StatusCode, string(data))
		}

		return queryResp, fmt.Errorf("not expected status code : %d", resp.StatusCode)
	}

	if err := json.Unmarshal(data, &queryResp); err != nil {
		return queryResp, errUnmarshalBody
	}

	return queryResp, nil
}
