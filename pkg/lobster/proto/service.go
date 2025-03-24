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

package proto

import (
	context "context"
	"errors"

	"github.com/naver/lobster/pkg/lobster/model"
	"github.com/naver/lobster/pkg/lobster/query"
	"github.com/naver/lobster/pkg/lobster/store"
	"github.com/naver/lobster/pkg/lobster/util"
)

type ChunkService struct {
	Store     *store.Store
	converter Converter
	UnimplementedChunkServiceServer
}

func (c *ChunkService) GetChunksWithinRange(ctx context.Context, req *Request) (*Response, error) {
	chunks, _ := c.Store.GetChunksWithinRange(query.Request{
		Start: util.Timestamp{Time: req.Start.AsTime()},
		End:   util.Timestamp{Time: req.End.AsTime()},
	})

	return &Response{
		ProtoChunk: c.converter.FromChunks(chunks),
	}, nil
}

func (c *ChunkService) GetChunk(ctx context.Context, req *Request) (*Response, error) {
	chunk := c.Store.LoadChunk(model.Source{
		Type: req.Source.Type,
		Path: req.Source.Path,
	}, req.PodUid, req.Container)

	if chunk == nil {
		return nil, errors.New("failed to load chunk")
	}

	return &Response{
		ProtoChunk: c.converter.FromChunks([]model.Chunk{*chunk}),
	}, nil
}
