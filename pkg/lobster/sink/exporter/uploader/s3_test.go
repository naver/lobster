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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/sink/order"
	v1 "github.com/naver/lobster/pkg/operator/api/v1"
)

func TestNewS3HTTPClientTimeouts(t *testing.T) {
	transport, ok := newS3HTTPClient().Transport.(*http.Transport)
	if !ok {
		t.Fatalf("unexpected transport type %T", newS3HTTPClient().Transport)
	}

	if transport.DialContext == nil {
		t.Error("expected custom dialer to be set")
	}
	if transport.TLSHandshakeTimeout != s3TLSHandshakeTimeout {
		t.Errorf("got TLS handshake timeout %v, want %v", transport.TLSHandshakeTimeout, s3TLSHandshakeTimeout)
	}
	if transport.ResponseHeaderTimeout != s3ResponseHeaderTimeout {
		t.Errorf("got response header timeout %v, want %v", transport.ResponseHeaderTimeout, s3ResponseHeaderTimeout)
	}
}

func TestS3UploadResponseHandling(t *testing.T) {
	const responseHeaderTimeout = 100 * time.Millisecond

	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
	}{
		{
			name: "ok",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.Copy(io.Discard, r.Body)
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
		{
			// A storage that never answers must not block the exporter forever.
			name: "stalled response times out",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.Copy(io.Discard, r.Body)
				<-r.Context().Done()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svr := httptest.NewServer(tt.handler)
			defer svr.Close()

			client := newS3HTTPClient()
			client.Transport.(*http.Transport).ResponseHeaderTimeout = responseHeaderTimeout

			uploader := S3Uploader{
				httpClient: client,
				Order: order.Order{
					SinkNamespace: "ns",
					SinkName:      "sink",
					LogExportRule: v1.LogExportRule{
						Name: "rule",
						S3Bucket: &v1.S3Bucket{
							Destination: svr.URL,
							BucketName:  "bucket",
							AccessKey:   "access",
							SecretKey:   "secret",
						},
					},
				},
			}

			now := time.Now()
			err := uploader.Upload([]byte("log line\n"), model.Chunk{}, now, now.Add(time.Minute))

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
