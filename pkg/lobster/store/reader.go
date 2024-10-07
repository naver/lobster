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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/logline"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/query/filter"
	"github.com/naver/lobster/pkg/lobster/util"
	"github.com/ncw/directio"
)

const BlockExt = ".log"

var (
	readerPool = sync.Pool{
		New: func() interface{} {
			return newBlockReader()
		},
	}
	largeBlockReaderPool = sync.Pool{
		New: func() interface{} {
			return newLargeBlockReader()
		},
	}
	readerBufferSize           = 16 * 1024
	blockBufferSize      int64 = 4 * 1024 * 1024  // 4mb
	largeBlockBufferSize int64 = 30 * 1024 * 1024 // 30mb
)

type blockReader struct {
	reader *bufio.Reader
	block  []byte
}

func newBlockReader() *blockReader {
	return &blockReader{bufio.NewReaderSize(nil, readerBufferSize), directio.AlignedBlock(int(blockBufferSize))}
}

func newLargeBlockReader() *blockReader {
	return &blockReader{bufio.NewReaderSize(nil, readerBufferSize), directio.AlignedBlock(int(largeBlockBufferSize))}
}

func loadBlocks(files []model.LogFile, conf config, blockFileFunc func(block model.ReadableBlock, checkPoint *model.CheckPoint, file model.LogFile)) {
	for _, file := range files {
		dir := fmt.Sprintf("%s/%s", *conf.StoreRootPath, file.RelativeBlockDir())
		cp, err := model.NewCheckPointFromFile(dir)
		if err != nil {
			glog.Errorf("failed to get cp %s : %s", err.Error(), dir)
			continue
		}

		if model.TempBlockFileName == file.FileName {
			block, err := loadTempBlock(fmt.Sprintf("%s/%s", dir, model.TempBlockFileName), cp.FileNum)
			if err != nil {
				glog.Errorf("failed to get temp block %s : %s", dir, err.Error())
				continue
			}
			blockFileFunc(block, cp, file)
			continue
		}

		block, err := fileToBlock(file)
		if err != nil {
			glog.V(3).Infof("failed to get block for %s : %s", file.Path, err.Error())
			continue
		}

		blockFileFunc(block, cp, file)
	}
}

func fileToBlock(file model.LogFile) (*model.Block, error) {
	tokens := strings.Split(strings.Replace(file.FileName, BlockExt, "", -1), "_")

	start, err := time.Parse(time.RFC3339Nano, tokens[0])
	if err != nil {
		return nil, err
	}

	end, err := time.Parse(time.RFC3339Nano, tokens[1])
	if err != nil {
		return nil, err
	}

	line, err := strconv.ParseInt(tokens[2], 0, 64)
	if err != nil {
		return nil, err
	}

	fileNum, err := strconv.ParseInt(tokens[3], 0, 64)
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(file.Path)
	if err != nil {
		return nil, err
	}

	return model.NewBlock(start, end, line, stat.Size(), fileNum), nil
}

func loadTempBlock(filePath string, fileNum int64) (*model.TempBlock, error) {
	var (
		start     time.Time
		end       time.Time
		line      = int64(0)
		size      = int64(0)
		now       = time.Now()
		blkReader *blockReader
	)

	f, err := directio.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cErr := f.Close(); err == nil {
			err = cErr
		}
	}()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if stat.Size() > blockBufferSize {
		blkReader = largeBlockReaderPool.Get().(*blockReader)
		defer largeBlockReaderPool.Put(blkReader)
	} else {
		blkReader = readerPool.Get().(*blockReader)
		defer readerPool.Put(blkReader)
	}

	numOfBytes, err := readFile(f, blkReader.block)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, err
	}

	reader := blkReader.reader
	reader.Reset(bytes.NewReader((blkReader.block)[:numOfBytes]))
	defer reader.Reset(nil)

	for {
		readBuffer, err := readBytes(reader, '\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		if len(readBuffer) < logline.MinTimestampLen {
			break // incompleted input (some logs are filed too fast)
		}

		ts, err := logline.ParseTimestamp(util.BytesToString(readBuffer))
		if err != nil {
			continue
		}

		if start.IsZero() {
			start = ts
		}

		if ts.After(now) {
			break
		}

		size = size + int64(len(readBuffer))
		line = line + 1
		end = ts
	}

	if end.IsZero() {
		return nil, fmt.Errorf("incompleted temp block")
	}

	return &model.TempBlock{StartedAt: start, EndedAt: end, Line: line, Size: size, FileNum: fileNum}, nil
}

func readBlocks(chunk model.Chunk, storeRootkDir string, onlySeries bool, start time.Time, end time.Time, filterers ...filter.Filterer) ([]byte, []model.Bucket, error) {
	buffer := &bytes.Buffer{}
	blocks := chunk.GetBlocksAfterTime(start)
	bucketBuilder := model.NewBucketBuilder(start, chunk)

	if start.IsZero() || end.IsZero() {
		return nil, []model.Bucket{}, errors.New("invalid range")
	}

	prevTs := time.Time{}
	for _, block := range blocks {
		if !(block.StartTime().Before(end) && block.EndTime().After(start)) {
			continue
		}

		skip, err := readBlock(chunk.Source.Type, block, fmt.Sprintf("%s/%s/%s", storeRootkDir, chunk.RelativeBlockDir, block.FileName()), onlySeries, buffer, bucketBuilder, prevTs, start, end, filterers...)
		if prevTs.Before(block.EndTime()) {
			prevTs = block.EndTime()
		}
		if skip {
			continue
		}
		if err != nil {
			return nil, []model.Bucket{}, err
		}
	}

	bucketBuilder.Save()

	return buffer.Bytes(), bucketBuilder.Build(), nil
}

func readBlock(sourceType string, block model.ReadableBlock, blockPath string, onlySeries bool, buffer *bytes.Buffer, bucketBuilder *model.BucketBuilder, prevTs, start, end time.Time, filterers ...filter.Filterer) (bool, error) {
	var blkReader *blockReader

	f, err := directio.OpenFile(blockPath, os.O_RDONLY, 0)
	if err != nil {
		glog.V(3).Infof("the block may have been removed by gc %s", blockPath)
		return true, nil
	}
	defer func() {
		err = errors.Join(err, errors.Join(f.Sync(), f.Close()))
	}()

	bucketBuilder.Reset(block.FileNumber(), block.StartTime())

	stat, err := f.Stat()
	if err != nil {
		glog.V(3).Infof("%v | failed to get stats %s", err, blockPath)
		return true, nil
	}

	if stat.Size() > blockBufferSize {
		blkReader = largeBlockReaderPool.Get().(*blockReader)
		defer largeBlockReaderPool.Put(blkReader)
	} else {
		blkReader = readerPool.Get().(*blockReader)
		defer readerPool.Put(blkReader)
	}

	numOfBytes, err := readFile(f, blkReader.block)
	if err != nil && err != io.ErrUnexpectedEOF {
		glog.Error(err)
		return true, nil
	}

	reader := blkReader.reader
	reader.Reset(bytes.NewReader(blkReader.block[:numOfBytes]))
	defer reader.Reset(nil)

	for {
		readBuffer, err := readBytes(reader, '\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return false, err
		}

		if len(readBuffer) < logline.MinTimestampLen {
			break // incompleted input (some logs are filed too fast)
		}

		ts, err := logline.ParseTimestamp(util.BytesToString(readBuffer))
		if err != nil {
			glog.Error(err)
			continue
		}

		if !prevTs.IsZero() && (ts.Before(prevTs) || ts.Equal(prevTs)) {
			continue // prevent to read duplicated contents
		}

		if ts.Before(start) {
			continue
		}

		if ts.After(end) {
			break
		}

		var msg string

		msg, err = logline.ParseLogMessageBySource(sourceType, util.BytesToString(readBuffer))
		if err != nil {
			glog.Error(err)
			continue
		}

		result, err := filter.DoFilter(msg, ts, filterers...)
		if err != nil {
			return false, err
		}

		if result == filter.Filtered {
			continue
		}

		if !bucketBuilder.IsWithinRange(ts) {
			bucketBuilder.Next(ts)
		}
		bucketBuilder.Pour(uint64(len(msg)))

		if result == filter.Done {
			break
		}

		if result == filter.SkipRead {
			continue
		}

		if onlySeries {
			continue
		}

		buffer.Write(readBuffer)
	}

	return false, nil
}

func readBytes(reader *bufio.Reader, delim byte) ([]byte, error) {
	buf, err := reader.ReadSlice(delim)
	if err != bufio.ErrBufferFull {
		return buf, err
	}

	front := make([]byte, len(buf))
	copy(front, buf)

	for {
		_, err = reader.ReadSlice(delim)
		if err != bufio.ErrBufferFull {
			return front, err
		}
	}
}

func readFile(r io.Reader, buf []byte) (n int, err error) {
	n, err = r.Read(buf)
	if n > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return
}
