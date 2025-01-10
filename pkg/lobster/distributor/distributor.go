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

package distributor

import (
	"io"
	"log"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/client"
	"github.com/naver/lobster/pkg/lobster/loader"
	"github.com/naver/lobster/pkg/lobster/logline"
	"github.com/naver/lobster/pkg/lobster/metrics"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/query"
	v1 "k8s.io/api/core/v1"

	"github.com/naver/lobster/pkg/lobster/sink/helper"
	"github.com/naver/lobster/pkg/lobster/sink/matcher"
	"github.com/naver/lobster/pkg/lobster/store"
	"github.com/naver/lobster/pkg/lobster/tailer"
	"github.com/naver/lobster/pkg/lobster/util"
)

var (
	conf config
)

type Distributor struct {
	tailerCache sync.Map
	store       *store.Store
	matcher     matcher.LogMatcher
	client      client.Client
}

func init() {
	conf = setup()
	log.Println("distributor configuration is loaded")
}

func NewDistributor(store *store.Store) Distributor {
	client, err := client.New()
	if err != nil {
		panic(err)
	}

	return Distributor{
		tailerCache: sync.Map{},
		store:       store,
		matcher:     matcher.NewLogMatcher(),
		client:      client,
	}
}

func (d *Distributor) Run(stopChan chan struct{}) {
	d.store.InitChunks()
	go func(stopChan chan struct{}) {
		inspectTicker := time.NewTicker(*conf.FileInspectInterval)

		defer func() {
			inspectTicker.Stop()
		}()

		for {
			select {
			case <-inspectTicker.C:
				podMap := d.client.GetPods()

				if len(podMap) == 0 {
					panic("no pods found")
				}

				d.updateLabelsInChunks(podMap)

				logfiles, err := d.loadLogFiles(podMap)
				if err != nil {
					glog.Error(err)
					continue
				}

				if len(logfiles) == 0 {
					panic("no log files found")
				}

				fileMap := d.extractFileMap(logfiles, *conf.FileInspectMaxStale)
				tailList := d.extractTailList(fileMap, *conf.TailFileMaxStale)

				if *conf.ShouldUpdateLogMatcher {
					now := time.Now()
					if err := d.matcher.Update(helper.FilterChunksByExistingPods(d.store.GetChunks(), podMap), now.Add(-*conf.FileInspectInterval), now); err != nil {
						glog.Error(err)
					}
				}

				d.storeFiles(fileMap)
				d.tailFiles(tailList, stopChan)

				d.store.Mark()
				d.store.Clean()

			case <-stopChan:
				glog.Info("stop distributor")
				return
			}
		}
	}(stopChan)
	go func(stopChan chan struct{}) {
		metricsTicker := time.NewTicker(*conf.MetricsInterval)

		defer func() {
			metricsTicker.Stop()
		}()

		for {
			select {
			case <-metricsTicker.C:
				end := time.Now()
				d.updateMetrics(end.Add(-*conf.MetricsInterval), end)
			case <-stopChan:
				glog.Info("stop metrics production")
				return
			}
		}
	}(stopChan)
}

func (d *Distributor) updateLabelsInChunks(podMap map[string]v1.Pod) {
	d.store.UpdateChunks(func(chunk *model.Chunk) {
		pod, ok := podMap[chunk.PodUID]

		if !ok {
			return
		}

		if reflect.DeepEqual(chunk.Labels, pod.Labels) {
			return
		}

		chunk.Labels = pod.Labels
		d.store.WriteLabelsFile(chunk)
	})
}

func (d *Distributor) loadLogFiles(podMap map[string]v1.Pod) ([]model.LogFile, error) {
	logFiles, err := loader.LoadLogfiles(*conf.StdstreamLogRootPath, func(podDirName string) (model.Labels, error) {
		return model.NewLabelsFromDirectoryName(podDirName, podMap)
	}, loader.ParseKubeLogFile)
	if err != nil {
		return nil, err
	}

	podLogfiles := loader.LoadPodEmptyDir(*conf.EmptyDirLogRootPath, podMap)
	if len(podLogfiles) > 0 {
		logFiles = append(logFiles, podLogfiles...)
	}

	return logFiles, nil
}

func (d *Distributor) extractFileMap(fileList []model.LogFile, maxStale time.Duration) map[string][]model.LogFile {
	fileMap := map[string][]model.LogFile{}
	list := fileList

	sort.Slice(list, func(i, k int) bool {
		return list[k].ModTime.After(list[i].ModTime)
	})

	for _, file := range list {
		if maxStale < time.Since(file.ModTime) {
			continue
		}

		fileMap[file.Id()] = append(fileMap[file.Id()], file)
	}

	return fileMap
}

func (d *Distributor) extractTailList(fileMap map[string][]model.LogFile, maxStale time.Duration) []model.LogFile {
	tailList := []model.LogFile{}

	for _, files := range fileMap {
		lastFile := files[len(files)-1]
		if maxStale < time.Since(lastFile.ModTime) {
			continue
		}

		tailList = append(tailList, lastFile)
	}

	return tailList
}

func (d *Distributor) storeFiles(fileMap map[string][]model.LogFile) {
	wg := sync.WaitGroup{}

	wg.Add(len(fileMap))

	for _, files := range fileMap {
		go func(files []model.LogFile) {
			defer wg.Done()

			if len(files) == 0 {
				return
			}

			var (
				chunk       *model.Chunk
				err         error
				targetFiles = files
				refFile     = files[0]
			)

			if d.store.HasChunk(refFile.Source, refFile.PodUID, refFile.Container) {
				newFiles := []model.LogFile{}
				chunk = d.store.LoadChunk(refFile.Source, refFile.PodUID, refFile.Container)

				for _, file := range files {
					if chunk.CheckPoint.FileNum < file.Number {
						newFiles = append(newFiles, file)
					}
				}

				if len(newFiles) == 0 {
					return
				}

				if v, loaded := d.tailerCache.LoadAndDelete(refFile.Id()); loaded {
					v.(*tailer.Tailer).Stop()
				}
				if err := d.store.MoveTempblock(chunk, chunk.CheckPoint.FileNum, refFile.Number); err != nil {
					glog.Error(err)
				}
				chunk.CheckPoint.Reset(refFile.Number)
				targetFiles = newFiles
			} else {
				chunk, err = model.NewChunk(files[0], nil)
				if err != nil {
					glog.Info(err)
					return
				}
			}

			d.store.WriteFiledLogs(chunk, targetFiles, d.handleMatches)
		}(files)
	}
	wg.Wait()
}

func (d *Distributor) tailFiles(fileList []model.LogFile, stopChan chan struct{}) {
	for _, file := range fileList {
		var err error

		key := file.Id()
		chunk := d.store.LoadChunk(file.Source, file.PodUID, file.Container)

		if chunk == nil {
			chunk, err = model.NewChunk(file, model.NewCheckPoint(file.Number, 0))
			if err != nil {
				continue
			}
			d.store.StoreChunk(file.Source, file.PodUID, file.Container, chunk)
		}

		if _, ok := d.tailerCache.Load(key); ok {
			glog.V(3).Infof("tailing already for %s", file.Path)
			continue
		}

		go d.storeTailedLogs(file, chunk, key, file.Number, stopChan)
	}
}

func (d *Distributor) storeTailedLogs(file model.LogFile, chunk *model.Chunk, key string, fileNum int64, stopChan chan struct{}) {
	defer func(key string) {
		if v, ok := d.tailerCache.Load(key); ok {
			v.(*tailer.Tailer).Stop()
			d.tailerCache.Delete(key)
		}
	}(key)

	if file.InspectedSize < chunk.CheckPoint.Offset {
		glog.V(3).Infof("truncated %s | file size: %d | offset: %d", file.Path, file.InspectedSize, chunk.CheckPoint.Offset)
		chunk.CheckPoint.Offset = 0

		// distinguish between new logs and old logs in temp block
		if err := d.store.MoveTempblock(chunk, chunk.CheckPoint.FileNum, chunk.CheckPoint.FileNum); err != nil {
			glog.Error(err)
		}
	}

	tail, err := tailer.NewTailer(file, chunk.CheckPoint.Offset, io.SeekStart)
	if err != nil {
		glog.Error(err)
		return
	}

	d.tailerCache.Store(key, tail)

	go tail.Run(stopChan)

	if err := d.store.WriteTailedLogs(chunk, fileNum, tail.LogChan, stopChan, d.handleMatches); err != nil {
		glog.V(3).Infof("%s | %s", err.Error(), key)
	}
}

func (d *Distributor) updateMetrics(start, end time.Time) {
	chunks, _ := d.store.GetChunksWithinRange(query.Request{
		Start: util.Timestamp{Time: start},
		End:   util.Timestamp{Time: end},
	})

	for _, chunk := range chunks {
		metrics.SetSizeOfBlocksInChunk(chunk.Namespace, chunk.Pod, chunk.Container, chunk.Source.Type, chunk.Source.Path, float64(chunk.Size))
	}

	limits := d.store.GetLimits()
	for _, limit := range limits {
		cap, used, _, _, description := limit.Stat()
		metrics.SetCapacityOfLimit(float64(cap), description)
		metrics.SetUsageOfLimit(float64(used), description)
	}
}

func (d *Distributor) handleMatches(chunk *model.Chunk, logLine string, logTs time.Time) {
	if time.Since(logTs) > *conf.MatchLookbackMin {
		return
	}

	body, err := logline.ParseLogMessageBySource(chunk.Source.Type, logLine)
	if err != nil {
		glog.Error(err)
		return
	}

	d.matcher.Match(chunk.Key(), body, logTs)
}
