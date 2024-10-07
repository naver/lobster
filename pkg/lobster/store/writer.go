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
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/logline"
	"github.com/naver/lobster/pkg/lobster/metrics"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/util"
)

func writeFiledLogs(chunk *model.Chunk, files []model.LogFile, blockDirPath string, maxBlockSize int64, logHandler LogHandler) ([]*model.Block, int64, error) {
	buf := emptyWriteBuffer()
	blocks := []*model.Block{}

	for _, file := range files {
		writtenBlocks, err := writeBlocks(chunk, file, buf, blockDirPath, maxBlockSize, logHandler)
		if err != nil {
			return blocks, 0, err
		}
		blocks = append(blocks, writtenBlocks...)
	}

	return blocks, buf.fileOffset, nil
}

func writeBlocks(chunk *model.Chunk, file model.LogFile, buf *writeBuffer, blockDirPath string, maxBlockSize int64, logHandler LogHandler) ([]*model.Block, error) {
	var (
		readLine string
		prevTs   time.Time
	)

	blocks := []*model.Block{}

	f, err := os.Open(file.Path)
	if err != nil {
		return blocks, err
	}

	defer func() {
		err = errors.Join(err, errors.Join(f.Sync(), f.Close()))
	}()

	reader := bufio.NewReader(f)
	buf.resetFileOffset()

	for {
		readLine, err = reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			return blocks, err
		}

		ts, err := logline.ParseTimestamp(readLine)
		if err != nil {
			glog.V(3).Info("failed to parse timestamp for %s: %s", file.Path, readLine)
			if file.Source.Type == model.LogTypeStdStream || prevTs.IsZero() {
				continue
			}

			ts = prevTs
			readLine = logline.MakeUnreliableTimestamp(ts, readLine)
		} else {
			prevTs = ts
		}

		if buf.start.IsZero() {
			buf.start = ts
		}

		buf.write(ts, readLine)

		go logHandler(chunk, readLine, ts)

		if int64(buf.size()) < maxBlockSize {
			continue
		}

		buf.end = ts
		block, err := writeBlock(blockDirPath, buf, file.Number)
		if err != nil {
			return blocks, err
		}
		if block != nil {
			blocks = append(blocks, block)
		}
		buf.reset()
		prevTs = time.Time{}
	}

	if buf.size() == 0 {
		return blocks, nil
	}

	buf.end = prevTs
	block, err := writeBlock(blockDirPath, buf, file.Number)
	if err != nil {
		return blocks, nil
	}
	if block != nil {
		blocks = append(blocks, block)
	}

	return blocks, nil
}

func writeBlock(dir string, buf *writeBuffer, fileNumber int64) (*model.Block, error) {
	var block *model.Block

	if buf.isValid() {
		return nil, fmt.Errorf("invalid timestamp order [%v - %v] %s", buf.start, buf.end, dir)
	}

	block = model.NewBlock(buf.start, buf.end, buf.lines, int64(buf.size()), fileNumber)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}
	filePath := fmt.Sprintf("%s/%s", dir, block.FileName())
	_, err := os.Stat(filePath)
	if err == nil { // skip if already exists
		return nil, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	if err := os.WriteFile(filePath, buf.bytes(), 0600); err != nil {
		return nil, err
	}
	return block, nil
}

func writeTailedLogs(chunk *model.Chunk, blockDirPath, tempBlockFilePath string, fileNum int64, maxBlockSize int64, logChan chan logline.LogLine, stopChan chan struct{}, bucket *leakyBucket, logHandler LogHandler) error {
	buf := emptyWriteBuffer()

	tempFile, err := os.OpenFile(tempBlockFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModeAppend)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, errors.Join(tempFile.Sync(), tempFile.Close()))
	}()

	flushTicker := time.NewTicker(bucket.interval)

	defer func() {
		flushTicker.Stop()
		if err := flushWriteBuffer(buf, tempFile, chunk, blockDirPath, fileNum, maxBlockSize); err != nil {
			glog.Error(err)
		}
	}()

	for {
		select {
		case line, ok := <-logChan:
			if !ok {
				return nil
			}

			if line.Err != nil {
				return line.Err
			}

			if line.Timestamp.IsZero() {
				continue
			}

			buf.lastOffset = line.Offset
			msg := line.Line + "\n"

			if buf.start.IsZero() {
				buf.start = line.Timestamp
			}
			buf.end = line.Timestamp

			buf.write(line.Timestamp, msg)

			go logHandler(chunk, msg, line.Timestamp)

			if ok, description := bucket.Pour(int64(len(msg))); !ok {
				buf.write(line.Timestamp, fmt.Sprintf("%s stdout F (lobsgter: Logs exceeding %s were limited)\n", line.Timestamp.Format(time.RFC3339Nano), description))
				metrics.AddOverloadedCount(chunk.Namespace, chunk.Pod, chunk.Container, chunk.Source.Type, chunk.Source.Path, description)

				return fmt.Errorf("logs are limit(%s) for %s_%s_%s", description, chunk.Namespace, chunk.Pod, chunk.Container)
			}

		case <-flushTicker.C:
			start := time.Now()
			if err := flushWriteBuffer(buf, tempFile, chunk, blockDirPath, fileNum, maxBlockSize); err != nil {
				return err
			}
			bucket.Init(start)
			metrics.ObserveFlushSeconds(chunk.Namespace, chunk.Pod, chunk.Container, chunk.Source.Type, chunk.Source.Path, time.Since(start).Seconds())
		case <-stopChan:
			return nil
		}
	}
}

func flushWriteBuffer(buf *writeBuffer, tempFile *os.File, chunk *model.Chunk, blockDirPath string, fileNum int64, maxBlockSize int64) error {
	defer buf.reset()

	if buf.size() == 0 {
		return nil
	}

	if buf.start.IsZero() {
		return fmt.Errorf("invalid wrtie buffer timestamp")
	}

	if chunk.TempBlock.StartedAt.IsZero() {
		chunk.TempBlock.StartedAt = buf.start
	}

	if _, err := tempFile.WriteString(buf.string()); err != nil {
		return err
	}

	chunk.CheckPoint.SetOffset(buf.lastOffset)
	if err := util.WriteFile(blockDirPath, model.CheckPointFileName, chunk.CheckPoint.ToBytes()); err != nil {
		return err
	}

	chunk.UpdateTempBlock(int64(buf.size()), buf.lines, buf.end)

	metrics.AddTailedBytes(chunk.Namespace, chunk.Pod, chunk.Container, chunk.Source.Type, chunk.Source.Path, float64(buf.size()))
	metrics.AddTailedLines(chunk.Namespace, chunk.Pod, chunk.Container, chunk.Source.Type, chunk.Source.Path, float64(buf.lines))

	if chunk.TempBlock.Size < maxBlockSize {
		return nil
	}

	if err := moveTempblock(chunk, blockDirPath, fileNum, fileNum); err != nil {
		return err
	}

	if _, err := tempFile.Seek(0, 0); err != nil {
		glog.Error(err)
	}

	return nil
}

func moveTempblock(chunk *model.Chunk, blockDirPath string, oldFileNum, newFileNum int64) error {
	newBlock := model.NewBlockFromTempBlock(*chunk.TempBlock, oldFileNum)

	tempBlockFilePath := fmt.Sprintf("%s/%s", blockDirPath, model.TempBlockFileName)
	newBlockFilePath := fmt.Sprintf("%s/%s", blockDirPath, newBlock.FileName())

	if err := util.MoveFileContents(tempBlockFilePath, newBlockFilePath); err != nil {
		return err
	}

	chunk.AppendBlocks([]*model.Block{newBlock})
	chunk.TempBlock.Reset(newFileNum)
	return nil
}

func setupBlockPathIfNotExist(dir string) error {
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	if err := makeFileIfNotExist(fmt.Sprintf("%s/%s", dir, model.TempBlockFileName)); err != nil {
		return err
	}

	if err := makeFileIfNotExist(fmt.Sprintf("%s/%s", dir, model.CheckPointFileName)); err != nil {
		return err
	}

	return nil
}

func makeFileIfNotExist(filePath string) error {
	_, err := os.Stat(filePath)
	if err == nil || !os.IsNotExist(err) {
		return err
	}
	newBlockFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if cErr := newBlockFile.Close(); err == nil {
			err = cErr
		}
	}()
	return nil
}
