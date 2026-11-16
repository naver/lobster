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

package server

import (
	"net"
	"testing"

	"github.com/naver/lobster/pkg/lobster/util"
)

func TestGetLocalEndpoint(t *testing.T) {
	port := *conf.ServerPort
	fallback := net.JoinHostPort(util.GetLocalAddress(), port)

	tests := []struct {
		name     string
		podIP    string
		expected string
	}{
		{name: "ipv4 pod ip", podIP: "10.1.2.3", expected: "10.1.2.3:" + port},
		{name: "ipv6 pod ip", podIP: "fd00::1", expected: "[fd00::1]:" + port},
		{name: "empty pod ip falls back to interface lookup", podIP: "", expected: fallback},
		{name: "invalid pod ip falls back to interface lookup", podIP: "not-an-ip", expected: fallback},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(envPodIP, tt.podIP)

			if got := (ApiServer{}).GetLocalEndpoint(); got != tt.expected {
				t.Errorf("GetLocalEndpoint() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
