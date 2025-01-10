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
	"bytes"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/sink/order"
	"github.com/naver/lobster/pkg/operator/api/v1/template"
	"github.com/pkg/errors"
)

var defaultRegion = aws.String("US")

type S3Uploader struct {
	Order order.Order
}

func NewS3Uploader(order order.Order) S3Uploader {
	return S3Uploader{
		Order: order,
	}
}

func (s S3Uploader) Type() string {
	return "S3"
}

func (s S3Uploader) Name() string {
	return s.Order.LogExportRule.Name
}

func (s S3Uploader) Interval() time.Duration {
	return s.Order.LogExportRule.Interval.Duration
}

func (s S3Uploader) Dir(chunk model.Chunk, date time.Time) string {
	if len(s.Order.LogExportRule.S3Bucket.PathTemplate) > 0 {
		path, err := s.templateDir(chunk, date)
		if err != nil {
			return s.defaultDir(chunk, date)
		}

		return path
	}

	return s.defaultDir(chunk, date)
}

func (b S3Uploader) FileName(start, end time.Time) string {
	fileName := fmt.Sprintf("%s_%s.log", start.Format(layoutFileName), end.Format(layoutFileName))

	if b.Order.LogExportRule.S3Bucket.ShouldEncodeFileName {
		return strings.ReplaceAll(fileName, "+", "%2B")
	}

	return fileName
}

func (s S3Uploader) Validate() error {
	return s.Order.LogExportRule.S3Bucket.Validate()
}

func (s S3Uploader) Upload(data []byte, dir, fileName string) error {
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

func (s S3Uploader) defaultDir(chunk model.Chunk, date time.Time) string {
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

func (s S3Uploader) templateDir(chunk model.Chunk, date time.Time) (string, error) {
	return template.GeneratePath(
		s.Order.LogExportRule.S3Bucket.PathTemplate,
		template.PathElement{
			Namespace:  chunk.Namespace,
			SinkName:   s.Order.SinkName,
			RuleName:   s.Order.LogExportRule.Name,
			Pod:        chunk.Pod,
			Container:  chunk.Container,
			SourceType: chunk.Source.Type,
			SourcePath: chunk.Source.Path,
			TimeInput:  date,
		})
}
