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
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/util"
)

const (
	SourceDelimeter = "::"
	OrDelimiter     = "|"
)

func ParseRequestWithUri(uri string) (Request, error) {
	req := newRequest()

	u, err := url.Parse(uri)
	if err != nil {
		return req, errors.New("invalid uri")
	}

	pId := u.Query().Get("id")
	pClusters := u.Query().Get("clusters")
	pNamespaces := u.Query().Get("namespaces")
	pLabels := u.Query().Get("labels")
	pSetNames := u.Query().Get("setNames")
	pPods := u.Query().Get("pods")
	pContainers := u.Query().Get("containers")
	pSources := u.Query().Get("sources")
	pNamespace := u.Query().Get("namespace")
	pSetName := u.Query().Get("setName")
	pPod := u.Query().Get("pod")
	pContainer := u.Query().Get("container")
	pStart := u.Query().Get("start")
	pEnd := u.Query().Get("end")
	pPage := u.Query().Get("page")
	pBurst := u.Query().Get("burst")
	pInclude := u.Query().Get("include")
	pExclude := u.Query().Get("exclude")

	req.Namespace = pNamespace
	req.SetName = pSetName
	req.Pod = pPod
	req.Container = pContainer

	if len(pSources) > 0 {
		sources := strings.Split(pSources, OrDelimiter)

		for _, source := range sources {
			part := strings.Split(source, SourceDelimeter)
			if len(part) == 2 {
				req.Sources = append(req.Sources, model.Source{Type: part[0], Path: part[1]})
			} else if len(part) == 1 {
				req.Sources = append(req.Sources, model.Source{Type: part[0]})
			} else {
				return req, errors.New("invalid source")
			}
		}
	}

	if len(pClusters) > 0 {
		req.Clusters = strings.Split(pClusters, OrDelimiter)
	}
	if len(pNamespaces) > 0 {
		req.Namespaces = strings.Split(pNamespaces, OrDelimiter)
	}
	if len(pSetNames) > 0 {
		req.SetNames = strings.Split(pSetNames, OrDelimiter)
	}
	if len(pPods) > 0 {
		req.Pods = strings.Split(pPods, OrDelimiter)
	}
	if len(pContainers) > 0 {
		req.Containers = strings.Split(pContainers, OrDelimiter)
	}

	if len(pLabels) > 0 {
		labels := []model.Labels{}
		pairs := strings.Split(pLabels, OrDelimiter)
		for _, pair := range pairs {
			kv := strings.Split(pair, model.LabelKeyValueDelimiter)
			if len(kv) == 2 {
				labels = append(labels, model.Labels{kv[0]: kv[1]})
			}
		}
		req.Labels = labels
	}

	if len(pId) == 0 {
		req.ID = uuid.New().String()
	} else {
		req.ID = pId
	}

	if len(pBurst) > 0 {
		burst, err := strconv.Atoi(pBurst)
		if err != nil {
			return req, errors.New("invalid parameter value `burst`")
		}

		req.Burst = burst
	}

	if len(pStart) != 0 || len(pEnd) != 0 {
		start, err := util.ConvertStringToTimestamp(pStart)
		if err != nil {
			return req, errors.New("invalid parameter value `start`")
		}
		end, err := util.ConvertStringToTimestamp(pEnd)
		if err != nil {
			return req, errors.New("invalid parameter value `end`")
		}

		req.Start = start
		req.End = end
	}

	if len(pInclude) != 0 {
		expr, err := url.QueryUnescape(pInclude)
		if err != nil {
			return req, errors.New("invalid parameter value `include`")
		}
		req.FilterIncludeExpr = expr
	}

	if len(pExclude) != 0 {
		expr, err := url.QueryUnescape(pExclude)
		if err != nil {
			return req, errors.New("invalid parameter value `exclude`")
		}
		req.FilterExcludeExpr = expr
	}

	if len(pPage) != 0 {
		page, err := strconv.Atoi(pPage)
		if err != nil {
			return req, errors.New("invalid parameter value `page`")
		}

		req.Page = page
	}

	return req, req.Init()
}
