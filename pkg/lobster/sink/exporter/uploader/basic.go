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
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/sink/order"
	v1 "github.com/naver/lobster/pkg/operator/api/v1"
	"github.com/naver/lobster/pkg/operator/api/v1/template"
)

type BasicUploader struct {
	httpClient *http.Client
	Order      order.Order
}

func NewBasicUploader(order order.Order) BasicUploader {
	return BasicUploader{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				IdleConnTimeout:     5 * time.Second,
				MaxIdleConns:        100,
				MaxConnsPerHost:     100,
				MaxIdleConnsPerHost: 100,
				Dial: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 5 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: time.Second,
			},
		},
		Order: order,
	}
}

func (b BasicUploader) Type() string {
	return "BasicBucket"
}

func (b BasicUploader) Name() string {
	return b.Order.LogExportRule.Name
}

func (b BasicUploader) Interval() time.Duration {
	return b.Order.LogExportRule.Interval.Duration
}

func (b BasicUploader) Dir(chunk model.Chunk, date time.Time) string {
	if len(b.Order.LogExportRule.BasicBucket.PathTemplate) > 0 {
		path, err := b.templateDir(chunk, date)
		if err != nil {
			return b.defaultDir(chunk, date)
		}

		return path
	}

	return b.defaultDir(chunk, date)
}

func (b BasicUploader) FileName(start, end time.Time) string {
	fileName := fmt.Sprintf("%s_%s.log", start.Format(layoutFileName), end.Format(layoutFileName))

	if b.Order.LogExportRule.BasicBucket.ShouldEncodeFileName {
		return strings.ReplaceAll(fileName, "+", "%2B")
	}

	return fileName
}

func (b BasicUploader) Validate() v1.ValidationErrors {
	return b.Order.LogExportRule.BasicBucket.Validate()
}

func (b BasicUploader) Upload(data []byte, dir, fileName string) error {
	var start = time.Now()

	u, err := url.Parse(b.Order.LogExportRule.BasicBucket.Destination)
	if err != nil {
		return err
	}

	u.Path = path.Join(u.Path, path.Join(dir, fileName))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	w, err := writer.CreateFormFile("text", fileName)
	if err != nil {
		return err
	}
	defer writer.Close()

	if _, err := w.Write(data); err != nil {
		return err
	}

	defer func() {
		glog.Infof("[basic][took %fs] upload %d bytes to %s", time.Since(start).Seconds(), len(data), u.String())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body.Bytes()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return err
	}

	var respBody []byte

	if resp != nil {
		respBody, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(respBody))
	}

	return err
}

func (b BasicUploader) defaultDir(chunk model.Chunk, date time.Time) string {
	dirPath := b.Order.Path()
	layout := b.Order.LogExportRule.BasicBucket.TimeLayoutOfSubDirectory
	rootPath := b.Order.LogExportRule.BasicBucket.RootPath

	if len(chunk.Source.Path) > 0 {
		dirPath = fmt.Sprintf("%s/%s", dirPath, chunk.Source.Path)
	}
	if len(layout) == 0 {
		layout = defaultLayout
	}
	if len(rootPath) == 0 {
		rootPath = "/"
	}

	return fmt.Sprintf("%s/%s/%s",
		rootPath,
		date.Format(layout),
		dirPath)
}

func (b BasicUploader) templateDir(chunk model.Chunk, date time.Time) (string, error) {
	return template.GeneratePath(
		b.Order.LogExportRule.BasicBucket.PathTemplate,
		template.PathElement{
			Cluster:    chunk.Cluster,
			Namespace:  chunk.Namespace,
			SinkName:   b.Order.SinkName,
			RuleName:   b.Order.LogExportRule.Name,
			Pod:        chunk.Pod,
			Container:  chunk.Container,
			SourceType: chunk.Source.Type,
			SourcePath: chunk.Source.Path,
			TimeInput:  date,
		})
}
