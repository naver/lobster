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

package db

import (
	bolt "go.etcd.io/bbolt"
)

type Database struct {
	db *bolt.DB
}

func NewDatabase(path string) *Database {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		panic(err)
	}
	return &Database{db: db}
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) GetOrCreate(bucketName []byte) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b != nil {
			return nil
		}
		_, err := tx.CreateBucket(bucketName)

		return err
	})
}

func (d *Database) Put(bucketName, key, value []byte) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(bucketName).Put(key, value)
		return err
	})
}

func (d *Database) Get(bucketName, key []byte) (data []byte, err error) {
	err = d.db.View(func(tx *bolt.Tx) error {
		data = tx.Bucket(bucketName).Get(key)
		return nil
	})
	return
}

func (d *Database) ForEach(bucketName []byte, iterFunc func(k, v []byte) error) error {
	return d.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketName).ForEach(iterFunc)
	})
}

func (d *Database) DeleteItems(bucketName []byte, itemKeys [][]byte) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		for _, key := range itemKeys {
			if err := tx.Bucket(bucketName).Delete(key); err != nil {
				return err
			}
		}
		return nil
	})
}
