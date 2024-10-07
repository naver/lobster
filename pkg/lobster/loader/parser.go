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

package loader

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/naver/lobster/pkg/lobster/model"
)

type ParseFunc func(logType, path string, mTime time.Time, size int64, appearance int) (*model.LogFile, error)

func ParseKubeLogFile(logType, path string, mTime time.Time, size int64, appearance int) (*model.LogFile, error) {
	subPaths := strings.Split(path, "/")
	if len(subPaths) == 0 {
		return nil, errors.New("invalid path : " + path)
	}

	tokens := strings.Split(subPaths[appearance], "_")
	if len(tokens) != 3 {
		return nil, errors.New("failed to parse path : " + path)
	}

	number, err := strconv.ParseInt(strings.Replace(subPaths[len(subPaths)-1], LogExt, "", -1), 0, 64)
	if err != nil {
		return nil, err
	}

	return &model.LogFile{
		Namespace: tokens[0],
		Pod:       tokens[1],
		PodUID:    tokens[2],
		Container: subPaths[appearance+1],
		FileName:  subPaths[appearance+2],
		Path:      path,
		Source: model.Source{
			Type: logType,
		},
		Number:        number,
		ModTime:       mTime,
		InspectedSize: size,
	}, nil
}

func ParseStoredLogFile(logType, path string, mTime time.Time, size int64, appearance int) (*model.LogFile, error) {
	subPaths := strings.Split(path, "/")
	if len(subPaths) == 0 {
		return nil, errors.New("invalid path : " + path)
	}

	tokens := strings.Split(subPaths[appearance], "_")
	if len(tokens) != 3 {
		return nil, errors.New("failed to parse path : " + path)
	}

	var (
		container       string
		pathInContainer string
		logDir          = subPaths[appearance+1]
	)

	switch logType {
	case model.LogTypeEmptyDirFile:
		subTokens := strings.Split(logDir, model.LogTypeDelimiter)
		if len(subTokens) != 2 {
			return nil, errors.New("failed to parse log type : " + logDir)
		}
		container = EmptyDirDescription
		pathInContainer = subTokens[1]
	default:
		container = logDir
	}

	fileName := subPaths[len(subPaths)-1]
	if fileName == model.TempBlockFileName {
		return &model.LogFile{
			Namespace: tokens[0],
			Pod:       tokens[1],
			PodUID:    tokens[2],
			Container: container,
			FileName:  model.TempBlockFileName,
			Path:      path,
			Source: model.Source{
				Type: logType,
				Path: pathInContainer,
			},
			ModTime:       mTime,
			InspectedSize: size,
		}, nil
	}

	parts := strings.Split(strings.Replace(fileName, LogExt, "", -1), "_")
	number, err := strconv.ParseInt(parts[len(parts)-1], 0, 64)
	if err != nil {
		return nil, err
	}

	return &model.LogFile{
		Namespace: tokens[0],
		Pod:       tokens[1],
		PodUID:    tokens[2],
		Container: container,
		FileName:  subPaths[appearance+2],
		Path:      path,
		Source: model.Source{
			Type: logType,
			Path: pathInContainer,
		},
		Number:        number,
		ModTime:       mTime,
		InspectedSize: size,
	}, nil
}
