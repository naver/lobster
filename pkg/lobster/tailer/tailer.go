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

package tailer

import (
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"github.com/golang/glog"

	"github.com/naver/lobster/pkg/lobster/logline"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/tailer/tail"
)

const (
	statusRunning status = 1
	statusIdle    status = 2
	drainWait            = time.Millisecond
)

type status int

type Tailer struct {
	tail    *tail.Tail
	file    model.LogFile
	ticker  *time.Ticker
	LogChan chan logline.LogLine
	pause   bool
	status  status
	once    sync.Once
}

var (
	ErrUnavailableTail = errors.New("failed to get line")
	ErrLimitedByBucket = errors.New("limited by bucket")

	conf config
)

func init() {
	conf = setup()
	log.Println("tailer *conf.iguration is loaded")
}

func NewTailer(file model.LogFile, offset int64, whence int) (*Tailer, error) {
	tailConfig := tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: true,
		Location: &tail.SeekInfo{
			Offset: offset,
			Whence: whence,
		},
		WaitTimeAfterRotation: *conf.WaitTimeAfterRotation,
		Logger:                tail.DiscardingLogger,
	}

	if *conf.ShowTailLog {
		tailConfig.Logger = log.New(os.Stdout, "[TAILER]", log.LstdFlags)
	}

	tf, err := tail.TailFile(file.Path, tailConfig)
	if err != nil {
		return &Tailer{}, err
	}

	return &Tailer{
		tail:    tf,
		file:    file,
		ticker:  time.NewTicker(*conf.TimeToLive),
		LogChan: make(chan logline.LogLine),
		pause:   false,
		status:  statusRunning,
	}, nil
}

func (t *Tailer) Run(stopChan chan struct{}) {
	var prevTs time.Time

	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
		}
		t.doStop()
	}()
	glog.V(3).Infof("Tailing %s", t.file.Path)
	for {
		if t.pause {
			return
		}
		select {
		case line, ok := <-t.tail.Lines:
			if !ok {
				t.LogChan <- logline.LogLine{Line: "", Err: ErrUnavailableTail}
				t.pause = true
				continue
			}

			if line.Err != nil {
				t.LogChan <- logline.LogLine{Line: "", Err: line.Err}
				t.pause = true
				continue
			}

			lineTs, err := logline.ParseTimestamp(line.Text)
			if err != nil {
				glog.V(3).Info("failed to parse timestamp for %s: %s", t.file.Path, line.Text)
				if t.file.Source.Type == model.LogTypeStdStream || prevTs.IsZero() {
					continue
				}
				lineTs = prevTs
				line.Text = logline.MakeUnreliableTimestamp(lineTs, line.Text)
			} else {
				prevTs = lineTs
			}

			if *conf.MinStaleTime < time.Since(lineTs) {
				// discard old
				continue
			}

			offset, err := t.tail.Tell()
			if err != nil {
				t.LogChan <- logline.LogLine{Timestamp: lineTs, Line: line.Text, Err: err}
				t.pause = true
				continue
			}

			t.LogChan <- logline.LogLine{Timestamp: lineTs, Line: line.Text, Offset: offset, Err: nil}
			t.status = statusRunning
		case <-t.ticker.C:
			if t.status == statusIdle {
				glog.V(3).Infof("stop to tail because of idle status for %s_%s", t.file.Pod, t.file.Container)
				t.pause = true
				continue
			}
			t.status = statusIdle
		case <-stopChan:
			t.pause = true
			continue
		}
	}
}

func (t *Tailer) Stop() {
	t.pause = true
	t.doStop()
}

func (t *Tailer) doStop() {
	t.once.Do(func() {
		t.ticker.Stop()
		t.drain()
		if err := t.tail.Stop(); err != nil {
			glog.Error(err)
		}
	})
}

func (t *Tailer) drain() {
	t.tail.Kill(nil)

	time.Sleep(drainWait)

	select {
	case <-t.LogChan:
	default:
		close(t.LogChan)
	}

	for t.tail.IsSending {
		select {
		case <-t.tail.Lines:
		default:
		}
	}

	select {
	case _, ok := <-t.tail.Lines:
		if ok {
			// If the drainLimit is exceeded, a goroutine leak may remain.
			// The limit is set to prevent excessive cpu usage.
			glog.Errorf("overlimit %s", t.file.Path)
		}
	default:
	}
}
