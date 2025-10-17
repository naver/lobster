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

package logline

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/util"
)

const unreliableTimestamp = "(lobster: Unreliable timstamp)"

var (
	conf       Config
	regexpTime = regexp.MustCompile("\"time\":\"(.+?)\"")
	regexpLog  = regexp.MustCompile("\"log\":\"(.+?)\"")

	MinTimestampLen   = len(time.RFC3339) - 5
	MaxTimestampLen   = len(time.RFC3339Nano)
	MaxMilliSecondLen = 9

	pool = sync.Pool{
		New: func() interface{} {
			return bufio.NewReaderSize(nil, readerBufferSize)
		},
	}
	poolSize         = 1000
	readerBufferSize = 16 * 1024
)

func init() {
	for i := 0; i < poolSize; i++ {
		pool.Put(pool.New())
	}
}

func MakeUnreliableTimestamp(ts time.Time, line string) string {
	return fmt.Sprintf("%s %s %s", ts.Format(time.RFC3339Nano), unreliableTimestamp, line)
}

func ParseLogMessageBySource(sourceType, str string) (string, error) {
	if sourceType == model.LogTypeStdStream {
		return ParseLogMessage(str)
	}

	return str, nil
}

func ParseLogMessage(str string) (string, error) {
	if len(str) == 0 {
		return "", errors.New("could not parse empty input")
	}

	switch *conf.LogFormat {
	case LogFormatText:
		return getLogMessageInTextLogLine(str)
	case LogFormatJson:
		return getLogMessageInJsonLogLine(str)
	default:
		return "", errors.New("unsupported logline.format")
	}
}

func ParseTimestamp(str string) (time.Time, error) {
	switch *conf.LogFormat {
	case LogFormatText:
		return getTimestampInTextLogLine(str)
	case LogFormatJson:
		return getTimestampInJsonLogLine(str)
	default:
		return time.Time{}, errors.New("unsupported logline.format")
	}
}

func ParseStream(str string) (string, error) {
	if len(str) == 0 {
		return "", errors.New("could not parse empty input")
	}

	if *conf.LogFormat != LogFormatText {
		return "", errors.New("unsupported logline.format")
	}

	streamLen := 6
	idx := strings.Index(str, "stdout")
	if idx <= 0 {
		idx = strings.Index(str, "stderr")
	}
	if idx <= 0 {
		return "", fmt.Errorf("can't find log message: %s", str)
	}

	return str[idx : idx+streamLen], nil
}

func ParseTag(str string) (string, error) {
	if len(str) == 0 {
		return "", errors.New("could not parse empty input")
	}

	if *conf.LogFormat != LogFormatText {
		return "", errors.New("unsupported logline.format")
	}

	tagLen := 1
	idx := strings.Index(str, "F")
	if idx <= 0 {
		idx = strings.Index(str, "P")
	}
	if idx <= 0 {
		return "", fmt.Errorf("can't find log message: %s", str)
	}

	return str[idx : idx+tagLen], nil
}

func getLogMessageInTextLogLine(str string) (string, error) {
	idx := strings.Index(str, "F")
	if idx <= 0 {
		idx = strings.Index(str, "P")
	}
	if idx <= 0 {
		return "", fmt.Errorf("can't find log message: %s", str)
	}

	logIdx := idx + 2
	if len(str) <= logIdx {
		return "", nil
	}

	return str[logIdx:], nil
}

func getLogMessageInJsonLogLine(str string) (string, error) {
	matches := regexpLog.FindStringSubmatch(str)
	if len(matches) < 2 {
		return "", fmt.Errorf("can't find timestamp: %s", str)
	}

	return matches[1], nil
}

func getTimestampInTextLogLine(str string) (time.Time, error) {
	if len(str) < MinTimestampLen || !(str[4] == '-' && str[7] == '-' && str[10] == 'T' && str[13] == ':' && str[16] == ':') {
		return time.Time{}, fmt.Errorf("could not parse improper input %s", str)
	}

	year := (((int(str[0])-'0')*10+int(str[1])-'0')*10+int(str[2])-'0')*10 + int(str[3]) - '0'
	month := time.Month((int(str[5])-'0')*10 + int(str[6]) - '0')
	day := (int(str[8])-'0')*10 + int(str[9]) - '0'
	hour := (int(str[11])-'0')*10 + int(str[12]) - '0'
	minute := (int(str[14])-'0')*10 + int(str[15]) - '0'
	second := (int(str[17])-'0')*10 + int(str[18]) - '0'

	ms := 0
	localOffsetStr := ""
	count := 0

	maxLengh := int(math.Min(float64(len(str)), float64(MaxTimestampLen)))

	for index := 19; index < maxLengh; index++ {
		if str[index] == '.' {
			continue
		}

		ms = ms * 10

		if str[index] == '	' || str[index] == ' ' || str[index] == 'Z' {
			break
		}

		if str[index] == '+' || str[index] == '-' {
			localOffsetStr = str[index : index+6]
			break
		}

		ms = ms + int(str[index]-'0')
		count = count + 1
	}

	if len(localOffsetStr) > 0 {
		tzHour, tzMinute := interpretLocalOffset(localOffsetStr)
		hour = hour + tzHour
		minute = minute + tzMinute
	}

	return time.Date(year, month, day, hour, minute, second, int(float64(ms)*math.Pow10(MaxMilliSecondLen-count-1)), time.UTC).Local(), nil
}

func interpretLocalOffset(offset string) (int, int) {
	op := 1
	colonOffset := 0

	if offset[0] == '+' {
		op = -1
	}
	if offset[3] == ':' {
		colonOffset = 1
	}

	return int(offset[1]-'0')*10 + int(offset[2]-'0')*op, int(offset[3+colonOffset]-'0')*10 + int(offset[4+colonOffset]-'0')*op
}

func getTimestampInJsonLogLine(str string) (time.Time, error) {
	matches := regexpTime.FindStringSubmatch(str)
	if len(matches) < 2 {
		return time.Time{}, fmt.Errorf("can't find timestamp: %s", str)
	}

	t, err := time.Parse(time.RFC3339Nano, matches[1])
	if err != nil {
		return time.Time{}, err
	}

	return t.Local(), nil
}

func ParseLogMessageTest(str, format string) (string, error) {
	if format == LogFormatText {
		return getLogMessageInTextLogLine(str)
	}
	return getLogMessageInJsonLogLine(str)
}

func ParseTimestampTest(str, format string) (time.Time, error) {
	switch format {
	case LogFormatText:
		return getTimestampInTextLogLine(str)
	case LogFormatJson:
		return getTimestampInJsonLogLine(str)
	}

	return time.Time{}, errors.New("unsupported logline.format")
}

func HasProperLogLine(f model.LogFile) bool {
	file, err := os.Open(f.Path)
	if err != nil {
		return false
	}
	defer func() {
		if err := file.Close(); err != nil {
			glog.Error(err)
		}
	}()

	reader := pool.Get().(*bufio.Reader)
	reader.Reset(file)

	defer func() {
		reader.Reset(nil)
		pool.Put(reader)
	}()

	buf, err := reader.ReadSlice('\n')
	if err != nil {
		return false
	}

	if len(buf) < MinTimestampLen {
		return false
	}

	_, err = ParseTimestamp(util.BytesToString(buf))
	return err == nil
}
