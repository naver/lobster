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

package counter

import (
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/golang/glog"
	"github.com/naver/lobster/pkg/lobster/sink/db"
)

const databaseFileName = "receipt.db"

var bucketName = []byte("receiptBucket")

type Counter struct {
	db *db.Database
}

func NewCounter(dataPath string) Counter {
	newDb := db.NewDatabase(filepath.Join(dataPath, databaseFileName))
	if err := newDb.GetOrCreate(bucketName); err != nil {
		panic(err)
	}
	return Counter{newDb}
}

func (c Counter) Produce(bytes int, exportTime time.Time, interval time.Duration, logTime time.Time) Receipt {
	return Receipt{
		ExportBytes: bytes,
		ExportTime:  exportTime,
		LogTime:     logTime,
	}
}

func (c Counter) Load(key string) (Receipt, bool, error) {
	receipt := Receipt{}
	data, err := c.db.Get(bucketName, []byte(key))

	if err != nil {
		return receipt, false, err
	}

	if len(data) == 0 {
		return receipt, false, nil
	}

	if err := json.Unmarshal(data, &receipt); err != nil {
		return receipt, false, err
	}

	return receipt, true, nil
}

func (c Counter) Store(key string, receipt Receipt) error {
	data, err := json.Marshal(receipt)
	if err != nil {
		return err
	}

	return c.db.Put(bucketName, []byte(key), data)
}

func (c Counter) Clean() {
	now := time.Now()
	targets := [][]byte{}

	if err := c.db.ForEach(bucketName, func(k, v []byte) error {
		receipt := Receipt{}

		if err := json.Unmarshal(v, &receipt); err != nil {
			glog.Error(err)
			return nil
		}

		if receipt.ExportInterval.Seconds() != 0 && !receipt.IsStale(now) {
			return nil
		}

		glog.Infof("delete stale receipts %s", string(k))
		targets = append(targets, k)

		return nil
	}); err != nil {
		glog.Error(err)
	}

	if err := c.db.DeleteItems(bucketName, targets); err != nil {
		glog.Error(err)
	}
}
