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

package store

import (
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/logline"
	"github.com/naver/lobster/pkg/lobster/metrics"

	"github.com/naver/lobster/pkg/lobster/loader"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/query"
	"github.com/naver/lobster/pkg/lobster/server/errors"
	"github.com/naver/lobster/pkg/lobster/util"
)

var conf config

type LogHandler func(chunk *model.Chunk, logLine string, logTs time.Time)

type Store struct {
	chunkCache          sync.Map
	lock                sync.RWMutex
	limitFuncs          []LimitFunc
	limiter             Limiter
	ReqMaxBurst         int64
	ReqCooldownDuration time.Duration
}

func init() {
	conf = setup()
	log.Println("store configuration is loaded")
}

func NewStore() (*Store, error) {
	if len(*conf.StoreRootPath) == 0 || *conf.RetentionSize <= 0 || *conf.RetentionTime <= 0 || *conf.BlockSize <= 0 {
		return nil, fmt.Errorf("invalid store arguments")
	}

	return &Store{
		chunkCache: sync.Map{},
		limitFuncs: []LimitFunc{
			LimitChunkSize(*conf.RetentionSize),
			LimitChunkTime(*conf.RetentionTime),
		},
		limiter:             NewLimiter(),
		ReqMaxBurst:         *conf.ReqMaxBurst,
		ReqCooldownDuration: *conf.ReqCooldownDuration,
	}, nil
}

func (s *Store) GetLimits() []*Limit {
	return s.limiter.GetLimits()
}

func (s *Store) GetStoreRootPath() *string {
	return conf.StoreRootPath
}

func (s *Store) SetStoreRootPath(path *string) {
	conf.StoreRootPath = path
}

func (s *Store) GetChunks() (chunks []model.Chunk) {
	s.chunkCache.Range(func(key, value interface{}) bool {
		chunk := value.(*model.Chunk)
		if chunk.HasBlocks() && !chunk.IsOutdated(*conf.RetentionTime) {
			chunks = append(chunks, *chunk)
		}
		return true
	})
	return
}

func (s *Store) GetChunksWithinRange(req query.Request) (chunks []model.Chunk, err error) {
	util.MeasureElapse(func() string {
		s.chunkCache.Range(func(key, value interface{}) bool {
			chunk := value.(*model.Chunk)
			if chunk.HasBlocks() && !chunk.IsOutdated(*conf.RetentionTime) && chunk.UpdatedAt.After(req.Start.Time) && chunk.StartedAt.Before(req.End.Time) {
				chunks = append(chunks, *chunk)
			}
			return true
		})

		return fmt.Sprintf("%d chunks", len(chunks))
	}, fmt.Sprintf("GetChunksWithinRange | %s ", req.String()))
	return
}

func (s *Store) GetSeriesInBlocksWithinRange(req query.Request) (numOfChunk int, series model.SeriesData, err error) {
	var buckets []model.Bucket

	s.lock.RLock()
	defer s.lock.RUnlock()

	util.MeasureElapse(func() string {
		chunk := s.LoadChunk(req.Source, req.PodUID, req.Container)

		if chunk == nil {
			return "no chunks"
		}

		_, buckets, err = readBlocks(*chunk, *conf.StoreRootPath, true, req.Start.Time, req.End.Time, req.Filterers...)
		if err != nil {
			glog.Error(err)
		}

		series = model.BucketsToSeries(req.Start.Time, req.End.Time, buckets)
		numOfChunk = 1

		return fmt.Sprintf("chunks %d | buckets %d | lines %d", numOfChunk, len(buckets), series.Lines())
	}, fmt.Sprintf("GetSeriesInBlocksWithinRange | %s ", req.String()))

	return
}

func (s *Store) GetBlocksWithinRange(req query.Request) (data []byte, numOfChunk int, pageInfo model.PageInfo, err error) {
	var buckets []model.Bucket

	s.lock.RLock()
	defer s.lock.RUnlock()

	util.MeasureElapse(func() string {
		chunk := s.LoadChunk(req.Source, req.PodUID, req.Container)

		if chunk == nil {
			return "no chunks"
		}

		data, buckets, err = readBlocks(*chunk, *conf.StoreRootPath, false, req.Start.Time, req.End.Time, req.Filterers...)
		if err != nil {
			glog.Error(err)
		}

		totalLines := int64(0)

		for _, bucket := range buckets {
			totalLines = totalLines + bucket.Lines
		}

		burst := req.Burst
		if burst == 0 {
			burst = *conf.PageBurst
		}

		pageInfo = model.NewPageInfo(req.Page, totalLines, int64(burst), false)
		numOfChunk = 1

		return fmt.Sprintf("chunks %d | data %d | buckets %d", numOfChunk, len(data), len(buckets))
	}, fmt.Sprintf("GetBlocksWithinRange | %s ", req.String()))

	return
}

func (s *Store) GetEntriesWithinRange(req query.Request) ([]model.Entry, int, model.PageInfo, error) {
	// do nothing
	return nil, 0, model.PageInfo{}, errors.ErrNotImplemented
}

func (s *Store) Validate(req query.Request) error {
	if len(req.PodUID) == 0 || len(req.Container) == 0 {
		return fmt.Errorf("invalid pod uid or container name")
	}

	return nil
}

func (s *Store) InitChunks() {
	files, err := loader.LoadLogfiles(*conf.StoreRootPath, func(podDirName string) (model.Labels, error) {
		return model.NewLabelsFromFile(fmt.Sprintf("%s/%s/%s", *conf.StoreRootPath, podDirName, model.LabelsFileName))
	}, loader.ParseStoredLogFile)
	if err != nil {
		glog.Error(err)
		return
	}

	loadBlocks(files, conf, func(block model.ReadableBlock, checkPoint *model.CheckPoint, file model.LogFile) {
		chunk := s.LoadChunk(file.Source, file.PodUID, file.Container)
		if chunk == nil {
			chunk, err = model.NewChunk(file, checkPoint)
			if err != nil {
				glog.Error(err)
				return
			}
			s.StoreChunk(file.Source, file.PodUID, file.Container, chunk)
		}

		tempBlock, ok := block.(*model.TempBlock)
		if ok {
			chunk.SetTempBlock(tempBlock)
		} else {
			chunk.AppendBlocks([]*model.Block{block.(*model.Block)})
		}
	})
}

func (s *Store) HasChunk(source model.Source, podUid, container string) bool {
	return s.LoadChunk(source, podUid, container) != nil
}

func (s *Store) StoreChunk(source model.Source, podUid, container string, chunk *model.Chunk) {
	s.chunkCache.Store(storeKey(podUid, container, source.String()), chunk)
}

func (s *Store) UpdateChunks(updateFn func(*model.Chunk)) {
	s.chunkCache.Range(func(key, value any) bool {
		updateFn(value.(*model.Chunk))
		return true
	})
}

func (s *Store) WriteLabelsFile(chunk *model.Chunk) {
	if err := util.WriteFile(s.podDirPath(*chunk), model.LabelsFileName, chunk.Labels.ToBytes()); err != nil {
		glog.V(3).Info("failed to wrtie labels in %s", s.podDirPath(*chunk))
	}
}

func (s *Store) LoadChunk(source model.Source, podUid, container string) *model.Chunk {
	chunk, ok := s.chunkCache.Load(storeKey(podUid, container, source.String()))
	if !ok {
		return nil
	}
	return chunk.(*model.Chunk)
}

func (s *Store) WriteFiledLogs(chunk *model.Chunk, files []model.LogFile, logHandler LogHandler) {
	var err error

	blocks, offset, err := writeFiledLogs(chunk, files, s.blockDirPath(*chunk), *conf.BlockSize, logHandler)
	if err != nil {
		glog.Error(err)
		return
	}
	if len(blocks) == 0 {
		glog.V(3).Infof("no blocks for %s\n", files[0].Path)
		return
	}

	chunk.SetCheckPoint(model.NewCheckPoint(files[len(files)-1].Number, offset))
	chunk.AppendBlocks(blocks)

	if err := util.WriteFile(s.blockDirPath(*chunk), model.CheckPointFileName, chunk.CheckPoint.ToBytes()); err != nil {
		glog.V(3).Info("failed to wrtie checkpoint: " + err.Error())
	}
	glog.V(3).Infof("add blocks %v for %s\n", len(blocks), chunk.RelativeBlockDir)
	s.StoreChunk(chunk.Source, chunk.PodUID, chunk.Container, chunk)
}

func (s *Store) WriteTailedLogs(chunk *model.Chunk, fileNum int64, logChan chan logline.LogLine, stopChan chan struct{}, logHandler LogHandler) error {
	blockDirPath := s.blockDirPath(*chunk)
	tempBlockFilePath := s.blockFilePath(*chunk, model.TempBlockFileName)

	if err := setupBlockPathIfNotExist(blockDirPath); err != nil {
		glog.Infof("chunk updated: %v | is outdated: %v | reason: %s", chunk.UpdatedAt, chunk.IsOutdated(*conf.RetentionTime), err.Error())
		return err
	}

	bucket := NewLeakyBucket(s.limiter, *conf.LeakyBucketInterval)
	defer bucket.Release()

	return writeTailedLogs(chunk, blockDirPath, tempBlockFilePath, fileNum, *conf.BlockSize, logChan, stopChan, bucket, logHandler)
}

func (s *Store) MoveTempblock(chunk *model.Chunk, oldFileNum, newFileNum int64) error {
	return moveTempblock(chunk, s.blockDirPath(*chunk), oldFileNum, newFileNum)
}

func (s *Store) Mark() {
	cap, used, err := util.DiskInfo(*conf.StoreRootPath)
	if err != nil {
		glog.Error(err)
		return
	}

	limit := float64(cap) * (*conf.SoftLimitRatioForDisk)

	metrics.SetDiskUsed(float64(used))
	metrics.SetDiskLimit(limit)

	if s.shouldMarkEntire(used, uint64(limit)) {
		s.markByEntireSize()
	} else {
		s.markByRetention()
	}
}

func (s *Store) Clear() {
	s.chunkCache.Range(func(key, value any) bool {
		s.chunkCache.Delete(key)
		return true
	})
}

func (s *Store) shouldMarkEntire(used, limit uint64) bool {
	return used > limit
}

func (s *Store) markByEntireSize() {
	s.chunkCache.Range(func(key, value interface{}) bool {
		chunk := value.(*model.Chunk)
		totalBlocks := len(chunk.Blocks)
		reductionIndex := int(math.Ceil(float64(totalBlocks) * (1 - *conf.SoftLimitRatioForBlocks)))

		for i := 0; i < reductionIndex; i++ {
			chunk.MarkBlockAt(i)
		}
		return true
	})
}

func (s *Store) markByRetention() {
	s.chunkCache.Range(func(key, value interface{}) bool {
		for _, f := range s.limitFuncs {
			f(value.(*model.Chunk))
		}
		return true
	})
}

func (s *Store) Clean() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cleanChunks()
	s.cleanDirectoriesWithEmptyBlocks()
}

func (s *Store) cleanChunks() {
	s.chunkCache.Range(func(key, value interface{}) bool {
		chunk := (value.(*model.Chunk))
		if chunk.DeletionMark {
			chunk.DeleteContainerFiles(*conf.StoreRootPath)
			s.chunkCache.Delete(key)
			metrics.Delete(chunk.Namespace, chunk.Pod, chunk.Container, chunk.Source.Type, chunk.Source.Path)
			metrics.DeleteMatchedLogs(chunk.Namespace, chunk.Pod, chunk.Container, chunk.Source.Type, chunk.Source.Path)
			glog.V(3).Infof("delete chunk : %s\n", key)
			return true
		}
		if chunk.DeletionMarkInBlock {
			tmp := chunk.Copy()
			offset := 0
			for i, block := range tmp {
				if block.DeletionMark {
					chunk.DeleteBlockAt(i-offset, *conf.StoreRootPath)
					offset = offset + 1
					glog.V(3).Infof("delete block : %s | [%v ~ %v]\n", block.FileName(), block.StartedAt, block.EndedAt)
				}
			}
			chunk.DeletionMarkInBlock = false
		}
		return true
	})
}

func (s *Store) cleanDirectoriesWithEmptyBlocks() {
	podFiles, err := os.ReadDir(*conf.StoreRootPath)
	if err != nil {
		glog.Error(err)
		return
	}

	for _, podFile := range podFiles {
		if !podFile.IsDir() {
			continue
		}

		podDir := fmt.Sprintf("%s/%s", *conf.StoreRootPath, podFile.Name())
		if hasNoBlocks, err := hasNoBlocksInPodDirectory(podDir); err != nil {
			glog.Error(err)
		} else if hasNoBlocks {
			os.RemoveAll(podDir)
		}
	}
}

func (s *Store) podDirPath(chunk model.Chunk) string {
	return fmt.Sprintf("%s/%s", *conf.StoreRootPath, chunk.RelativePodDir)
}

func (s *Store) blockDirPath(chunk model.Chunk) string {
	return fmt.Sprintf("%s/%s", *conf.StoreRootPath, chunk.RelativeBlockDir)
}

func (s *Store) blockFilePath(chunk model.Chunk, blockFileName string) string {
	return fmt.Sprintf("%s/%s", s.blockDirPath(chunk), blockFileName)
}

func storeKey(podUid, container, source string) string {
	return fmt.Sprintf("%s_%s_%s", podUid, container, source)
}

func hasNoBlocksInPodDirectory(podDir string) (bool, error) {
	n, err := util.CountSubDirectories(podDir)
	if err != nil {
		return false, err
	}

	return n == 0, nil
}
