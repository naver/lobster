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

package model

import (
	"encoding/json"
	"testing"
	"time"

	gojson "github.com/goccy/go-json"
)

// createTestChunk creates a test chunk for benchmarking
func createTestChunk() Chunk {
	return Chunk{
		Id:        "test-chunk-123",
		Cluster:   "production-cluster",
		Namespace: "default",
		Labels: Labels{
			"app":         "web-server",
			"environment": "production",
			"version":     "v1.2.3",
			"team":        "platform",
		},
		SetName:   "web-deployment",
		Pod:       "web-server-abc123-xyz789",
		PodUid:    "550e8400-e29b-41d4-a716-446655440000",
		Container: "nginx",
		Source: Source{
			Type: "kubernetes",
			Path: "/var/log/pods/default_web-server-abc123-xyz789_550e8400/nginx/0.log",
		},
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
		Line:      10000,
		Size:      1024000,
	}
}

var (
	testChunk     = createTestChunk()
	testTimestamp = time.Now()
	testMessages  = []string{
		"[INFO] 2024-01-21 10:15:30 - Server started successfully on port 8080",
		"[ERROR] 2024-01-21 10:15:31 - Failed to connect to database: connection timeout after 30s",
		"[WARN] 2024-01-21 10:15:32 - High memory usage detected: 85% of available memory in use",
		"[DEBUG] 2024-01-21 10:15:33 - Processing request: GET /api/v1/users?page=1&limit=100",
		`[INFO] 2024-01-21 10:15:34 - JSON response: {"status":"ok","data":{"users":[{"id":1,"name":"John"}]}}`,
	}
)

// BenchmarkEntry_StdlibJSON benchmarks using standard library encoding/json
func BenchmarkEntry_StdlibJSON(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := testMessages[i%len(testMessages)]
		entry := NewEntry(testTimestamp, testChunk, msg)
		_, err := json.Marshal(entry)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEntry_GoccyJSON benchmarks using goccy/go-json library
func BenchmarkEntry_GoccyJSON(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := testMessages[i%len(testMessages)]
		entry := NewEntry(testTimestamp, testChunk, msg)
		_, err := gojson.Marshal(entry)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEntry_GoccyJSON_Parallel benchmarks parallel JSON marshaling (real-world scenario)
func BenchmarkEntry_GoccyJSON_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			msg := testMessages[i%len(testMessages)]
			entry := NewEntry(testTimestamp, testChunk, msg)
			_, err := gojson.Marshal(entry)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// BenchmarkEntry_ByMessageSize benchmarks performance with different message sizes
func BenchmarkEntry_ByMessageSize(b *testing.B) {
	sizes := []struct {
		name string
		msg  string
	}{
		{"Short", "Error occurred"},
		{"Medium", testMessages[0]},
		{"Long", testMessages[4]},
		{"VeryLong", string(make([]byte, 1024))}, // 1KB message
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				entry := NewEntry(testTimestamp, testChunk, size.msg)
				_, err := gojson.Marshal(entry)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkEntry_Creation measures the overhead of Entry creation only
func BenchmarkEntry_Creation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := testMessages[i%len(testMessages)]
		_ = NewEntry(testTimestamp, testChunk, msg)
	}
}

// BenchmarkEntry_FullPipeline benchmarks the complete pipeline (creation + marshaling + append)
func BenchmarkEntry_FullPipeline(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := testMessages[i%len(testMessages)]
		entry := NewEntry(testTimestamp, testChunk, msg)
		data, err := gojson.Marshal(entry)
		if err != nil {
			b.Fatal(err)
		}
		// In actual usage: buffer.Write(ts, append(data, '\n'))
		_ = append(data, '\n')
	}
}

// BenchmarkEntry_ByLabelsSize benchmarks performance with different numbers of labels
func BenchmarkEntry_ByLabelsSize(b *testing.B) {
	createChunkWithLabels := func(n int) Chunk {
		chunk := testChunk
		chunk.Labels = make(Labels)
		for i := 0; i < n; i++ {
			chunk.Labels[string(rune('a'+i))] = "value"
		}
		return chunk
	}

	sizes := []int{0, 5, 10, 20, 50}
	for _, size := range sizes {
		b.Run(string(rune('0'+size/10))+string(rune('0'+size%10))+"_labels", func(b *testing.B) {
			chunk := createChunkWithLabels(size)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				entry := NewEntry(testTimestamp, chunk, testMessages[0])
				_, err := gojson.Marshal(entry)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkEntry_Allocations measures memory allocation patterns
func BenchmarkEntry_Allocations(b *testing.B) {
	b.Run("WithAllocation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			entry := NewEntry(testTimestamp, testChunk, testMessages[0])
			data, _ := gojson.Marshal(entry)
			result := append(data, '\n')
			_ = result
		}
	})

	b.Run("PreallocatedBuffer", func(b *testing.B) {
		b.ReportAllocs()
		buf := make([]byte, 0, 1024)
		for i := 0; i < b.N; i++ {
			entry := NewEntry(testTimestamp, testChunk, testMessages[0])
			data, _ := gojson.Marshal(entry)
			buf = append(buf[:0], data...)
			buf = append(buf, '\n')
			_ = buf
		}
	})
}
