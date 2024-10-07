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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/logline"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/util"
	v1 "k8s.io/api/core/v1"
)

const (
	LogExt              = ".log"
	EmptyDirDescription = "__emptydir__"
)

type LabelFunc func(string) (model.Labels, error)

func LoadLogfiles(path string, labelFunc LabelFunc, parseFunc ParseFunc) ([]model.LogFile, error) {
	logfiles := []model.LogFile{}
	appearance := util.TargetPathAppearance(path)
	podFiles, err := os.ReadDir(path)
	if err != nil {
		return logfiles, err
	}

	for _, pf := range podFiles {
		if !pf.IsDir() {
			continue
		}

		logfilesOfPod := []model.LogFile{}
		podDir := fmt.Sprintf("%s/%s", path, pf.Name())
		subDirs, err := os.ReadDir(podDir)
		if err != nil {
			glog.Error(err)
			continue
		}

		for _, sd := range subDirs {
			if !sd.IsDir() {
				continue
			}

			files := loadLogfilesInDirectory(getLogType(sd.Name()), fmt.Sprintf("%s/%s", podDir, sd.Name()), appearance, parseFunc)
			if len(files) > 0 {
				logfilesOfPod = append(logfilesOfPod, files...)
			}
		}

		labels, err := labelFunc(pf.Name())
		if err != nil && !os.IsNotExist(err) {
			glog.Error(err)
		}

		for i := 0; i < len(logfilesOfPod); i++ {
			logfilesOfPod[i].Labels = labels
		}

		if len(logfilesOfPod) > 0 {
			logfiles = append(logfiles, logfilesOfPod...)
		}
	}

	return logfiles, nil
}

func loadLogfilesInDirectory(logType, dir string, appearance int, parseFunc ParseFunc) []model.LogFile {
	logfiles := []model.LogFile{}
	files, err := os.ReadDir(dir)
	if err != nil {
		glog.Error(err)
		return logfiles
	}

	for _, lf := range files {
		info, err := lf.Info()
		if err != nil {
			continue
		}
		if filepath.Ext(lf.Name()) != LogExt || info.Size() == 0 {
			continue
		}

		path := fmt.Sprintf("%s/%s", dir, lf.Name())
		modTime := info.ModTime()
		size := info.Size()

		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			originFile, err := os.Readlink(path)
			if err != nil {
				glog.Error(err)
				continue
			}
			stat, err := os.Stat(originFile)
			if err != nil {
				glog.Error(err)
				continue
			}
			modTime = stat.ModTime()
			size = stat.Size()
		}

		f, err := parseFunc(logType, path, modTime, size, appearance)
		if err != nil {
			glog.Errorf("%s | %s", err.Error(), path)
			continue
		}

		logfiles = append(logfiles, *f)
	}

	return logfiles
}

func LoadPodEmptyDir(root string, podMap map[string]v1.Pod) []model.LogFile {
	logfiles := []model.LogFile{}

	for uid, pod := range podMap {
		empyDirPath := emptyDir(root, uid)
		files := findLogFiles(empyDirPath)

		for path, file := range files {
			logfile := model.LogFile{
				Namespace: pod.Namespace,
				Labels:    pod.Labels,
				Pod:       pod.Name,
				PodUID:    uid,
				Container: EmptyDirDescription,
				FileName:  file.Name(),
				Path:      path,
				Source: model.Source{
					Type: model.LogTypeEmptyDirFile,
					Path: sanitizePath(path, empyDirPath),
				},
				Number:        0,
				ModTime:       file.ModTime(),
				InspectedSize: file.Size(),
			}

			if !logline.HasProperLogLine(logfile) {
				continue
			}

			logfiles = append(logfiles, logfile)
		}
	}

	return logfiles
}

func findLogFiles(root string) map[string]os.FileInfo {
	files := map[string]os.FileInfo{}

	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if filepath.Ext(path) != LogExt {
			return nil
		}

		files[path] = info

		return nil
	}); err != nil {
		glog.Error(err)
	}

	return files
}

func sanitizePath(path, cutPrefix string) string {
	return strings.ReplaceAll(strings.ReplaceAll(path, cutPrefix, ""), "/", "_")
}

func emptyDir(root, podUid string) string {
	return fmt.Sprintf("%s/%s/volumes/kubernetes.io~empty-dir", root, podUid)
}

func getLogType(dir string) string {
	if strings.HasPrefix(dir, model.LogTypeEmptyDirFile) {
		return model.LogTypeEmptyDirFile
	}

	return model.LogTypeStdStream
}
