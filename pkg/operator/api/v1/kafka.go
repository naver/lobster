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

package v1

import (
	"fmt"
)

const PartitionAny = -1

type TLS struct {
	// Whether or not to use TLS
	Enable bool `json:"enable,omitempty"`
	// Whether or not to skip verification of CA certificate in client
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
	// CA certificate for TLS
	CaCertificate string `json:"caCertificate,omitempty"`
}

type SASL struct {
	// Whether or not to use SASL authentication
	Enable bool `json:"enable,omitempty"`
	// Enabled SASL mechanism
	Mechanism string `json:"mechanism,omitempty"`
	// OAuth access token
	AccessToken string `json:"accessToken,omitempty"`
	// SASL Protocol Version
	Version int16 `json:"version,omitempty"`
	// Kafka SASL handshake
	Handshake bool `json:"handshake,omitempty"`
	// SASL/PLAIN or SASL/SCRAM authentication
	User string `json:"user,omitempty"`
	// Password for SASL/PLAIN authentication
	Password string `json:"password,omitempty"`
}

type Kafka struct {
	// Target kafka broker servers to send logs
	Brokers []string `json:"brokers,omitempty"`
	// TLS configuration
	TLS TLS `json:"tls,omitempty"`
	// SASL configuration
	SASL SASL `json:"sasl,omitempty"`
	// An identifier to distinguish request; default `lobster`
	ClientId string `json:"clientId,omitempty"`
	// Target topic to which logs will be exported (required)
	Topic string `json:"topic"`
	// Target partition to which logs will be exported (optional)
	Partition int32 `json:"partition,omitempty"`
	// Target key to which logs will be exported (optional)
	Key string `json:"key,omitempty"`
}

func (k Kafka) Validate() error {
	if len(k.Brokers) == 0 {
		return fmt.Errorf("`brokers` should not be empty")
	}

	if k.TLS.Enable && !k.TLS.InsecureSkipVerify && len(k.TLS.CaCertificate) == 0 {
		return fmt.Errorf("`caCertificate` should not be empty when TLS is enabled")
	}

	if len(k.Topic) == 0 {
		return fmt.Errorf("`topic` should not be empty")
	}

	if k.Partition < PartitionAny {
		return fmt.Errorf("`partition` should not be less than PartitionAny(-1)")
	}

	return nil
}
