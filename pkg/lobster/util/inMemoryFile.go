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

package util

import (
	"io"
	"os"
	"time"
)

type MemFile struct {
	Name string
	pos  int
	data []byte
}

func NewMemFile(name string, data []byte) *MemFile {
	return &MemFile{
		Name: name,
		pos:  0,
		data: data,
	}
}

func (f *MemFile) Read(p []byte) (int, error) {
	if f.pos >= len(f.data) {
		return 0, io.EOF
	}

	n := copy(p, f.data[f.pos:])
	f.pos += n

	return n, nil
}

func (f *MemFile) Seek(offset int64, whence int) (int64, error) {
	npos := f.pos

	switch whence {
	case io.SeekStart:
		npos = int(offset)
	case io.SeekCurrent:
		npos += int(offset)
	case io.SeekEnd:
		npos = len(f.data) + int(offset)
	default:
		npos = -1
	}

	if npos < 0 {
		return 0, os.ErrInvalid
	}

	f.pos = npos

	return int64(f.pos), nil
}

func (f *MemFile) Stat() (os.FileInfo, error) {
	return &MemFileInfo{f}, nil
}

func (f MemFile) Close() error {
	return nil
}

type MemFileInfo struct {
	file *MemFile
}

func (i *MemFileInfo) Name() string {
	return i.file.Name
}

func (i *MemFileInfo) Size() int64 {
	return int64(len(i.file.data))
}

func (s *MemFileInfo) Mode() os.FileMode {
	return os.ModeTemporary
}

func (s *MemFileInfo) ModTime() time.Time {
	return time.Now()
}

func (s *MemFileInfo) IsDir() bool {
	return false
}

func (s *MemFileInfo) Sys() interface{} {
	return nil
}
