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
	"github.com/naver/lobster/pkg/lobster/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Converter struct{}

func (c Converter) FromChunks(chunks []model.Chunk) (protoChunks []*ProtoChunk) {
	for _, chunk := range chunks {
		protoChunks = append(protoChunks, &ProtoChunk{
			Id:        chunk.Id,
			Cluster:   chunk.Cluster,
			Namespace: chunk.Namespace,
			Labels:    chunk.Labels,
			SetName:   chunk.SetName,
			Pod:       chunk.Pod,
			PodUid:    chunk.PodUid,
			Container: chunk.Container,
			Source: &ProtoSource{
				Type: chunk.Source.Type,
				Path: chunk.Source.Path,
			},
			Blocks: c.fromBlocks(chunk.Blocks),
			TempBlock: &ProtoTempBlock{
				StartedAt: timestamppb.New(chunk.TempBlock.StartedAt),
				EndedAt:   timestamppb.New(chunk.TempBlock.EndedAt),
				Line:      chunk.TempBlock.Line,
				Size:      chunk.TempBlock.Size,
				FileNum:   chunk.TempBlock.FileNum,
			},
			StartedAt:        timestamppb.New(chunk.StartedAt),
			UpdatedAt:        timestamppb.New(chunk.UpdatedAt),
			RelativePodDir:   chunk.RelativePodDir,
			Line:             chunk.Line,
			Size:             chunk.Size,
			RelativeBlockDir: chunk.RelativeBlockDir,
		})
	}

	return
}

func (c Converter) ToChunks(protoChunks []*ProtoChunk) (chunks []model.Chunk) {
	for _, chunk := range protoChunks {
		chunks = append(chunks, model.Chunk{
			Id:        chunk.Id,
			Cluster:   chunk.Cluster,
			Namespace: chunk.Namespace,
			Labels:    chunk.Labels,
			SetName:   chunk.SetName,
			Pod:       chunk.Pod,
			PodUid:    chunk.PodUid,
			Container: chunk.Container,
			Source: model.Source{
				Type: chunk.Source.Type,
				Path: chunk.Source.Path,
			},
			Blocks: c.toBlocks(chunk.Blocks),
			TempBlock: &model.TempBlock{
				StartedAt: chunk.TempBlock.StartedAt.AsTime(),
				EndedAt:   chunk.TempBlock.EndedAt.AsTime(),
				Line:      chunk.TempBlock.Line,
				Size:      chunk.TempBlock.Size,
				FileNum:   chunk.TempBlock.FileNum,
			},
			StartedAt:        chunk.StartedAt.AsTime(),
			UpdatedAt:        chunk.UpdatedAt.AsTime(),
			RelativePodDir:   chunk.RelativePodDir,
			Line:             chunk.Line,
			Size:             chunk.Size,
			RelativeBlockDir: chunk.RelativeBlockDir,
		})
	}

	return
}

func (c Converter) fromBlocks(blocks []*model.Block) (protoBlocks []*ProtoBlock) {
	for _, block := range blocks {
		protoBlocks = append(protoBlocks, &ProtoBlock{
			StartedAt: timestamppb.New(block.StartedAt),
			EndedAt:   timestamppb.New(block.EndedAt),
			Line:      block.Line,
			Size:      block.Size,
			FileNum:   block.FileNum,
		})
	}

	return
}

func (c Converter) toBlocks(protoBlocks []*ProtoBlock) (blocks []*model.Block) {
	for _, block := range protoBlocks {
		blocks = append(blocks, &model.Block{
			StartedAt: block.StartedAt.AsTime(),
			EndedAt:   block.EndedAt.AsTime(),
			Line:      block.Line,
			Size:      block.Size,
			FileNum:   block.FileNum,
		})
	}

	return
}
