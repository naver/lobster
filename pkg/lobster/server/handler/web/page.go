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

package web

import (
	"fmt"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/query"
)

var tHome *template.Template

type page struct {
	Namespaces        map[string]struct{}
	Clusters          map[string]struct{}
	Labels            map[string]struct{}
	SetNames          map[string]struct{}
	Pods              map[string]struct{}
	Containers        map[string]struct{}
	Sources           map[string]struct{}
	TotalPage         int
	IsPartialContents bool
	Contents          []byte
	HistogramScript   string
}

func init() {
	if _, err := os.Stat("web/home.html"); err != nil {
		panic(err)
	}
	tHome = template.Must(template.New("home.html").Funcs(template.FuncMap{
		"btoa": func(b []byte) string { return string(b) },
		"ttoi": func(t time.Time) int64 { return t.Unix() },
	}).ParseFiles("web/home.html"))
}

func newPage() page {
	return page{
		Namespaces: map[string]struct{}{},
		Clusters:   map[string]struct{}{},
		Labels:     map[string]struct{}{},
		SetNames:   map[string]struct{}{},
		Pods:       map[string]struct{}{},
		Containers: map[string]struct{}{},
		Sources:    map[string]struct{}{},
	}
}

func (p *page) fillPanel(chunks []model.Chunk) {
	for _, chunk := range chunks {
		p.Namespaces[chunk.Namespace] = struct{}{}
		for _, pair := range chunk.Labels.Pairs() {
			p.Labels[pair] = struct{}{}
		}
		p.Clusters[chunk.Cluster] = struct{}{}
		p.SetNames[chunk.SetName] = struct{}{}
		p.Pods[chunk.Pod] = struct{}{}
		p.Containers[chunk.Container] = struct{}{}
		p.Sources[makeSourceName(chunk.Source)] = struct{}{}
	}
}

func (p *page) fillContents(contents []byte, seriesData model.SeriesData, pageInfo model.PageInfo) (err error) {
	p.Contents = contents
	p.TotalPage = pageInfo.Total
	p.IsPartialContents = pageInfo.IsPartialContents
	p.HistogramScript, err = executeTemplate(seriesData)

	return
}

func (p *page) render(w http.ResponseWriter) error {
	return tHome.Execute(w, p)
}

func makeSourceName(source model.Source) string {
	switch source.Type {
	case model.LogTypeStdStream:
		return model.LogTypeStdStream
	default:
		return fmt.Sprintf("%s%s%s", source.Type, query.SourceDelimeter, source.Path)
	}
}
