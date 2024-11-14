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

package model

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
)

const LabelsFileName = "labels"

type Labels map[string]string

func (l Labels) Pairs() []string {
	kvs := []string{}

	for k, v := range l {
		kvs = append(kvs, fmt.Sprintf("%s%s%s", k, LabelKeyValueDelimiter, v))
	}

	return kvs
}

func (l Labels) PairKeyMap() map[string]bool {
	pairKeyMap := map[string]bool{}

	for k, v := range l {
		pairKeyMap[fmt.Sprintf("%s%s%s", k, LabelKeyValueDelimiter, v)] = true
	}

	return pairKeyMap
}

func (l Labels) String() string {
	return strings.Join(l.Pairs(), LabelsDelimiter)
}

func NewLabelsFromFile(filePath string) (Labels, error) {
	labels := Labels{}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return labels, err
	}

	if err := json.Unmarshal(data, &labels); err != nil {
		return labels, err
	}

	return labels, nil
}

func NewLabelsFromDirectoryName(dirName string, podMap map[string]v1.Pod) (Labels, error) {
	tokens := strings.Split(dirName, "_")
	if pod, ok := podMap[tokens[2]]; ok {
		return pod.Labels, nil
	}

	return Labels{}, fmt.Errorf("failed to get labels from directory %s", dirName)
}

func (d Labels) ToBytes() []byte {
	data, err := json.Marshal(d)
	if err != nil {
		glog.Error(err)
	}
	return data
}
