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
	"fmt"
	"time"
)

type LogFile struct {
	Namespace     string            `json:"namespace"`
	Labels        map[string]string `json:"labels"`
	Pod           string            `json:"pod"`
	PodUID        string            `json:"podUid"`
	Container     string            `json:"container"`
	FileName      string            `json:"fileName"`
	Path          string            `json:"path"`
	Source        Source            `json:"source"`
	Number        int64             `json:"number"`
	ModTime       time.Time         `json:"modTime"`
	InspectedSize int64             `json:"inspectedSize"`
}

func (f LogFile) RelativePodDir() string {
	return fmt.Sprintf("%s_%s_%s", f.Namespace, f.Pod, f.PodUID)
}

func (f LogFile) RelativeBlockDir() string {
	switch f.Source.Type {
	case LogTypeEmptyDirFile:
		return fmt.Sprintf("%s/%s%s%s", f.RelativePodDir(), f.Source.Type, LogTypeDelimiter, f.Source.Path)
	default:
		return fmt.Sprintf("%s/%s", f.RelativePodDir(), f.Container)
	}
}

func (f LogFile) Id() string {
	return f.RelativeBlockDir()
}
