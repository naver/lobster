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

package bucket

import (
	"bytes"
	"fmt"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/sink/order"
	"github.com/pkg/errors"
)

var defaultRegion = aws.String("US")

type S3Bucket struct {
	Order          order.Order
	LastUploadTime time.Time
}

func NewS3Bucket(order order.Order) S3Bucket {
	return S3Bucket{
		Order:          order,
		LastUploadTime: time.Now(),
	}
}

func (s S3Bucket) Type() string {
	return s.Order.SinkType
}

func (s S3Bucket) Name() string {
	return s.Order.LogExportRule.Name
}

func (s S3Bucket) Interval() time.Duration {
	return s.Order.LogExportRule.Interval.Duration
}

func (s S3Bucket) Dir(chunk model.Chunk, date time.Time) string {
	dirPath := s.Order.Path()
	layout := s.Order.LogExportRule.S3Bucket.TimeLayoutOfSubDirectory

	if len(chunk.Source.Path) > 0 {
		dirPath = fmt.Sprintf("%s/%s", dirPath, chunk.Source.Path)
	}
	if len(layout) == 0 {
		layout = defaultLayout
	}

	return fmt.Sprintf("%s/%s/%s",
		s.Order.LogExportRule.S3Bucket.RootPath,
		date.Format(layout),
		dirPath)
}

func (s S3Bucket) FileName(start, end time.Time) string {
	return fmt.Sprintf("%s_%s.log", start.Format(layoutFileName), end.Format(layoutFileName))
}

func (s S3Bucket) Validate() error {
	return s.Order.LogExportRule.S3Bucket.Validate()
}

func (s S3Bucket) Flush(data []byte, dir, fileName string) error {
	s3Session, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(s.Order.LogExportRule.S3Bucket.Destination),
		Credentials:      credentials.NewStaticCredentials(s.Order.LogExportRule.S3Bucket.AccessKey, s.Order.LogExportRule.S3Bucket.SecretKey, ""),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
		Region:           defaultRegion,
	})
	if err != nil {
		return err
	}

	if len(s.Order.LogExportRule.S3Bucket.Region) > 0 {
		s3Session.Config.Region = aws.String(s.Order.LogExportRule.S3Bucket.Region)
	}

	input := &s3manager.UploadInput{
		Bucket: aws.String(s.Order.LogExportRule.S3Bucket.BucketName),
		Key:    aws.String(path.Join(dir, fileName)),
		Body:   bytes.NewReader(data),
	}

	if len(s.Order.LogExportRule.S3Bucket.Tags) > 0 {
		input.Tagging = aws.String(s.Order.LogExportRule.S3Bucket.Tags.String())
	}

	result, err := s3manager.NewUploader(s3Session).Upload(input)
	if err != nil {
		return errors.Wrap(err, "failed to upload file")
	}

	glog.Infof("[s3] upload %d bytes to %s", len(data), result.Location)

	return nil
}
