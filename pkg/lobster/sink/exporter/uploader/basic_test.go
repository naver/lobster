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

package uploader

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/sink/order"
	v1 "github.com/naver/lobster/pkg/operator/api/v1"
)

func TestBasicUploadStatusHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		respBody   string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "ok",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			// A file for the same range already exists; retrying would fail forever,
			// so the upload is skipped instead of reported as a failure.
			name:       "conflict is skipped",
			statusCode: http.StatusConflict,
			respBody:   "already exists",
			wantErr:    false,
		},
		{
			name:       "server error is reported",
			statusCode: http.StatusInternalServerError,
			respBody:   "boom",
			wantErr:    true,
			wantErrMsg: "unexpected status code 500: boom",
		},
		{
			name:       "created is not treated as success",
			statusCode: http.StatusCreated,
			wantErr:    true,
			wantErrMsg: "unexpected status code 201",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.respBody))
			}))
			defer svr.Close()

			uploader := NewBasicUploader(order.Order{
				SinkNamespace: "ns",
				SinkName:      "sink",
				LogExportRule: v1.LogExportRule{
					Name:        "rule",
					BasicBucket: &v1.BasicBucket{Destination: svr.URL},
				},
			})

			now := time.Now()
			err := uploader.Upload([]byte("log line\n"), model.Chunk{}, now, now.Add(time.Minute))

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("got error %q, want it to contain %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}
