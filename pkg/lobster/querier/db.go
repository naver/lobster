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

package querier

import (
	"time"

	"github.com/hashicorp/go-memdb"
	"github.com/naver/lobster/pkg/lobster/model"
)

const (
	tableName           = "chunk"
	indexID             = "id"
	indexChunkNamespace = "chunk_namespace"
	indexStoreAddr      = "store_addr"
)

type Database struct {
	db *memdb.MemDB
}

func NewDatabase() (Database, error) {
	db, err := memdb.NewMemDB(&memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			tableName: {
				Name: tableName,
				Indexes: map[string]*memdb.IndexSchema{
					indexID: {
						Name:    indexID,
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Id"},
					},
					indexChunkNamespace: {
						Name:   indexChunkNamespace,
						Unique: false,
						Indexer: &memdb.CompoundMultiIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "Namespace"},
							},
						},
					},
					indexStoreAddr: {
						Name:    indexStoreAddr,
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "StoreAddr"},
					},
				},
			},
		},
	})
	return Database{
		db: db,
	}, err
}

func (d Database) getChunks() ([]model.Chunk, error) {
	var chunks []model.Chunk

	txn := d.db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get(tableName, indexID)
	if err != nil {
		return chunks, err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		chunks = append(chunks, obj.(model.Chunk))
	}

	return chunks, nil
}

func (d Database) getAllChunksWithinRange(start, end time.Time) ([]model.Chunk, error) {
	txn := d.db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get(tableName, indexID)
	if err != nil {
		return nil, err
	}

	return getChunksWithinRange(it, start, end), nil
}

func (d Database) getChunksForNamespaceWithinRange(namespace string, start, end time.Time) ([]model.Chunk, error) {
	txn := d.db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get(tableName, indexChunkNamespace, namespace)
	if err != nil {
		return nil, err
	}

	return getChunksWithinRange(it, start, end), nil
}

func getChunksWithinRange(it memdb.ResultIterator, start, end time.Time) []model.Chunk {
	var chunks []model.Chunk

	for obj := it.Next(); obj != nil; obj = it.Next() {
		chunk := obj.(model.Chunk)

		if chunk.UpdatedAt.After(start) && chunk.StartedAt.Before(end) {
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}

func (d Database) insert(chunks []model.Chunk) error {
	txn := d.db.Txn(true)

	for _, chunk := range chunks {
		val := chunk
		if err := txn.Insert(tableName, val); err != nil {
			return err
		}
	}
	txn.Commit()
	return nil
}

func (d Database) delete(chunk model.Chunk) error {
	txn := d.db.Txn(true)

	if err := txn.Delete(tableName, chunk); err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func (d Database) deleteByAddr(addr string) error {
	txn := d.db.Txn(true)
	if _, err := txn.DeleteAll(tableName, indexStoreAddr, addr); err != nil {
		return err
	}
	txn.Commit()
	return nil
}
