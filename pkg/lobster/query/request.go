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

package query

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/naver/lobster/pkg/lobster/loader"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/query/filter"
	"github.com/naver/lobster/pkg/lobster/util"
	sinkV1 "github.com/naver/lobster/pkg/operator/api/v1"
)

type Request struct {
	// Use internally
	ID string `json:"id,omitempty"`
	// Get chunks belongs to clusters
	Clusters []string `json:"clusters,omitempty"`
	// Get chunks belongs to namespaces
	Namespaces []string `json:"namespaces,omitempty"`
	// Get chunks belongs to namespaces and labels
	Labels []model.Labels `json:"labels,omitempty"`
	// Get chunks belongs to namespace and set names(replicaset/statefulset)
	SetNames []string `json:"setNames,omitempty"`
	// Get chunks belongs to namespace and pods
	Pods []string `json:"pods,omitempty"`
	// Get chunks belongs to namespace and containers
	Containers []string `json:"containers,omitempty"`
	// Get chunks belongs to namespace and log source
	Sources []model.Source `json:"sources,omitempty"`

	// Use internally
	Namespace string `json:"namespace,omitempty"`
	// Use internally
	SetName string `json:"setName,omitempty"`
	// Use internally
	Pod string `json:"pod,omitempty"`
	// Use internally
	PodUid string `json:"pod_uid,omitempty"`
	// Use internally
	Container string `json:"container,omitempty"`
	// Use internally
	Source model.Source `json:"source,omitempty"`

	// Start time for query
	Start util.Timestamp `json:"start,omitempty" swaggertype:"string"`
	// End time for query
	End util.Timestamp `json:"end,omitempty" swaggertype:"string"`
	// The page number for the returned logs
	Page int `json:"page,omitempty"`
	// The number of logs that can be returned in one page and this can be greater or less than burst
	Burst int `json:"burst,omitempty"`
	// Regular expression to search logs
	FilterIncludeExpr string            `json:"include,omitempty"`
	FilterExcludeExpr string            `json:"exclude,omitempty"`
	Filterers         []filter.Filterer `json:"-"`
	// Use internally
	Local         bool   `json:"local,omitempty" default:"false"`
	Attachment    bool   `json:"attachment,omitempty" default:"false"`
	Version       string `json:"-"`
	ContentsLimit uint64 `json:"-"`
}

func ParseRequestWithBody(body []byte) (Request, error) {
	req := newRequest()

	if err := json.Unmarshal(body, &req); err != nil {
		return req, err
	}

	req.Filterers = []filter.Filterer{}

	return req, req.Init()
}

func newRequest() Request {
	return Request{Filterers: []filter.Filterer{}}
}

func NewRequestFromFilter(f sinkV1.Filter) Request {
	labels := []model.Labels{}
	for _, label := range f.Labels {
		labels = append(labels, model.Labels(label))
	}

	sources := []model.Source{}
	for _, source := range f.Sources {
		sources = append(sources, model.Source{Type: source.Type, Path: source.Path})
	}

	return Request{
		Clusters:          f.Clusters,
		Namespace:         f.Namespace,
		Labels:            labels,
		SetNames:          f.SetNames,
		Pods:              f.Pods,
		Sources:           sources,
		Containers:        f.Containers,
		FilterIncludeExpr: f.FilterIncludeExpr,
		FilterExcludeExpr: f.FilterExcludeExpr,
		Filterers:         []filter.Filterer{},
	}
}

func (r *Request) Init() error {
	if err := r.InitRangeFilterer(); err != nil {
		return err
	}

	if err := r.InitTextFilterer(); err != nil {
		return err
	}

	return r.InitSource()
}

func (r *Request) InitRangeFilterer() error {
	if r.Start.Time.IsZero() || r.End.Time.IsZero() {
		return errors.New("invalid parameter value `start` or `end`")
	}

	r.Filterers = append(r.Filterers, filter.NewRangeFilter(r.Start.Time, r.End.Time))

	return nil
}

func (r *Request) InitTextFilterer() error {
	if len(r.FilterIncludeExpr) > 0 {
		filter, err := filter.NewRegexpFilterer(r.FilterIncludeExpr)
		if err != nil {
			return err
		}

		r.Filterers = append(r.Filterers, filter)
	}

	if len(r.FilterExcludeExpr) > 0 {
		filter, err := filter.NewNegativeRegexpFilterer(r.FilterExcludeExpr)
		if err != nil {
			return err
		}

		r.Filterers = append(r.Filterers, filter)
	}

	return nil
}

func (r *Request) InitSource() error {
	if r.Container == loader.EmptyDirDescription || strings.HasPrefix(r.Source.Path, model.LogTypeEmptyDirFile) {
		r.Source.Type = model.LogTypeEmptyDirFile
	}

	return nil
}

func (r Request) HasSetNames() bool {
	return len(r.Namespaces) != 0 && len(r.SetNames) != 0
}

func (r Request) HasPods() bool {
	return len(r.Namespaces) != 0 && len(r.Pods) != 0
}

func (r Request) HasContainers() bool {
	return len(r.Namespaces) != 0 && len(r.Containers) != 0
}

func (r Request) HasNamespace() bool {
	return len(r.Namespace) != 0
}

func (r Request) HasNamespaces() bool {
	return len(r.Namespaces) != 0
}

func (r Request) HasLabels() bool {
	return len(r.Labels) > 0
}

func (r Request) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}
