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
	"time"
)

type Entry struct {
	Timestamp  time.Time         `json:"time"`
	SourceType string            `json:"sourceType"`
	SourcePath string            `json:"sourcePath"`
	Stream     string            `json:"stream"`
	Tag        string            `json:"tag"`
	Cluster    string            `json:"cluster"`
	Namespace  string            `json:"namespace"`
	Labels     map[string]string `json:"labels"`
	Pod        string            `json:"pod"`
	Container  string            `json:"container"`
	Message    string            `json:"message"`
}

func NewEntry(c Chunk) Entry {
	return Entry{
		SourceType: c.Source.Type,
		SourcePath: c.Source.Path,
		Cluster:    c.Cluster,
		Namespace:  c.Namespace,
		Pod:        c.Pod,
		Container:  c.Container,
		Labels:     c.Labels,
	}
}

func (e Entry) String() string {
	return e.Message
}
