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
	"crypto/tls"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/sink/order"
	"github.com/naver/lobster/pkg/lobster/util"
	v1 "github.com/naver/lobster/pkg/operator/api/v1"
)

const (
	defaultClientId = "lobster"
	dialTimeout     = time.Second
)

type TokenProvider struct {
	token string
}

// In OAuth authentication, the required credentials
// for authentication vary across different systems
// Therefore, Lobster does not perform OAuth authentication
// directly but instead relies on the user to provide the access token.
func (t TokenProvider) Token() (*sarama.AccessToken, error) {
	return &sarama.AccessToken{Token: t.token}, nil
}

type KafkaUploader struct {
	Order order.Order
}

func NewKafkaUploader(order order.Order) KafkaUploader {
	return KafkaUploader{
		Order: order,
	}
}

func (k KafkaUploader) Type() string {
	return "Kafka"
}

func (k KafkaUploader) Name() string {
	return k.Order.LogExportRule.Name
}

func (k KafkaUploader) Interval() time.Duration {
	return k.Order.LogExportRule.Interval.Duration
}

func (k KafkaUploader) Dir(chunk model.Chunk, date time.Time) string {
	return ""
}

func (k KafkaUploader) FileName(start, end time.Time) string {
	return ""
}

func (k KafkaUploader) Validate() error {
	return k.Order.LogExportRule.Kafka.Validate()
}

func (k KafkaUploader) Upload(data []byte, dir, fileName string) error {
	config, err := newConfig(k.Order.LogExportRule.Kafka)
	if err != nil {
		return err
	}

	producer, err := sarama.NewSyncProducer(k.Order.LogExportRule.Kafka.Brokers, config)
	if err != nil {
		return err
	}
	defer producer.Close()

	if err := producer.SendMessages(newMessages(k.Order.LogExportRule.Kafka, data)); err != nil {
		return err
	}

	glog.Infof("[kafka] upload %d bytes to topic `%s` for %s", len(data), k.Order.LogExportRule.Kafka.Topic, k.Order.Request.String())

	return nil
}

func newConfig(kafka *v1.Kafka) (*sarama.Config, error) {
	config := sarama.NewConfig()

	config.ClientID = defaultClientId
	config.Producer.Return.Successes = true
	config.Net.DialTimeout = dialTimeout

	if kafka.TLS.Enable {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = &tls.Config{
			InsecureSkipVerify: kafka.TLS.InsecureSkipVerify,
		}

		if len(kafka.TLS.CaCertificate) > 0 {
			pool, err := util.NewCertPoolForRootCA([]byte(kafka.TLS.CaCertificate))
			if err != nil {
				return nil, err
			}

			config.Net.TLS.Config.RootCAs = pool
		}
	}

	if kafka.SASL.Enable {
		config.Net.SASL.Enable = true
		config.Net.SASL.Version = kafka.SASL.Version
		config.Net.SASL.Handshake = kafka.SASL.Handshake
		config.Net.SASL.Mechanism = sarama.SASLMechanism(kafka.SASL.Mechanism)

		switch config.Net.SASL.Mechanism {
		case sarama.SASLTypePlaintext:
			config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
			config.Net.SASL.User = kafka.SASL.User
			config.Net.SASL.Password = kafka.SASL.Password

		case sarama.SASLTypeSCRAMSHA256:
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
			config.Net.SASL.User = kafka.SASL.User
			config.Net.SASL.Password = kafka.SASL.Password
			config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &util.XDGSCRAMClient{HashGeneratorFcn: util.SHA256} }

		case sarama.SASLTypeSCRAMSHA512:
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
			config.Net.SASL.User = kafka.SASL.User
			config.Net.SASL.Password = kafka.SASL.Password
			config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &util.XDGSCRAMClient{HashGeneratorFcn: util.SHA512} }

		case sarama.SASLTypeOAuth:
			config.Net.SASL.Mechanism = sarama.SASLTypeOAuth
			config.Net.SASL.TokenProvider = TokenProvider{kafka.SASL.AccessToken}

		default:
			return nil, fmt.Errorf("Unsupported SASL mechanism: " + kafka.SASL.Mechanism)
		}
	}

	if len(kafka.ClientId) > 0 {
		config.ClientID = kafka.ClientId
	}

	return config, nil
}

func newMessages(kafka *v1.Kafka, data []byte) []*sarama.ProducerMessage {
	var (
		start    int
		index    int
		b        byte
		messages = []*sarama.ProducerMessage{}
	)

	for index, b = range data {
		if b != '\n' && index < len(data)-1 {
			continue
		}

		messages = append(messages, newMessage(kafka, data[start:index]))
		start = index + 1
	}

	return messages
}

func newMessage(kafka *v1.Kafka, data []byte) *sarama.ProducerMessage {
	message := &sarama.ProducerMessage{
		Topic:     kafka.Topic,
		Partition: kafka.Partition,
		Value:     sarama.ByteEncoder(data),
	}

	if len(kafka.Key) > 0 {
		message.Key = sarama.StringEncoder(kafka.Key)
	}

	return message
}
