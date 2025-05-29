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
	"github.com/IBM/sarama"
)

const (
	PartitionAny = -1

	OAuthTypeUnencodedCredential = "UnencodedCredential"
	OAuthTypeAuthenzPrincipal    = "AuthenzPrincipal"
)

type OAuthType string

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
	// SASL Protocol Version
	Version int16 `json:"version,omitempty"`
	// Kafka SASL handshake
	Handshake bool `json:"handshake,omitempty"`

	// SASL/PLAIN or SASL/SCRAM authentication
	User string `json:"user,omitempty"`
	// Password for SASL/PLAIN authentication
	Password string `json:"password,omitempty"`

	// Deprecated; OAuth access token
	AccessToken string `json:"accessToken,omitempty"`
	// Application's ID
	ClientID string `json:"clientId,omitempty"`
	// Application's secret
	ClientSecret string `json:"clientSecret,omitempty"`
	// TokenURL server endpoint to obtain the access token
	TokenURL string `json:"tokenUrl,omitempty"`
	// Scopes used to specify permission
	Scopes []string `json:"scopes,omitempty"`
	// Type for reflecting authentication server's specific requirements
	OAuthType OAuthType `json:"oAuthType,omitempty"`
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
	// the producer will ensure that exactly one
	Idempotent bool `json:"idempotent,omitempty"`
}

func (k Kafka) Validate() ValidationErrors {
	var validationErrors ValidationErrors

	if len(k.Brokers) == 0 {
		validationErrors.AppendErrorWithFields("kafka.brokers", ErrorEmptyField)
	}

	if k.TLS.Enable && !k.TLS.InsecureSkipVerify && len(k.TLS.CaCertificate) == 0 {
		validationErrors.AppendErrorWithFields("kafka.tls.caCertificate", ErrorEmptyField)
	}

	if k.SASL.Enable {
		switch k.SASL.Mechanism {
		case sarama.SASLTypeOAuth:
			if len(k.SASL.ClientID) == 0 {
				validationErrors.AppendErrorWithFields("kafka.sasl.clientId", ErrorEmptyField)
			}
			if len(k.SASL.ClientSecret) == 0 {
				validationErrors.AppendErrorWithFields("kafka.sasl.clientSecret", ErrorEmptyField)
			}
			if len(k.SASL.TokenURL) == 0 {
				validationErrors.AppendErrorWithFields("kafka.sasl.tokenUrl", ErrorEmptyField)
			}
		case sarama.SASLTypePlaintext:
			fallthrough
		case sarama.SASLTypeSCRAMSHA256:
			fallthrough
		case sarama.SASLTypeSCRAMSHA512:
			if len(k.SASL.User) == 0 {
				validationErrors.AppendErrorWithFields("kafka.sasl.user", ErrorEmptyField)
			}
			if len(k.SASL.Password) == 0 {
				validationErrors.AppendErrorWithFields("kafka.sasl.password", ErrorEmptyField)
			}
		default:
			validationErrors.AppendErrorWithFields("kafka.sasl.mechanism", "unsupported sasl auth mechanism")
		}
	}

	if len(k.Topic) == 0 {
		validationErrors.AppendErrorWithFields("kafka.topic", ErrorEmptyField)
	}

	if k.Partition < PartitionAny {
		validationErrors.AppendErrorWithFields("kafka.partition", "the partition value must not be less than PartitionAny(-1)")
	}

	return validationErrors
}
