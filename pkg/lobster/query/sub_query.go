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

package query

import (
	"fmt"
	"time"

	"github.com/naver/lobster/pkg/lobster/model"
	svrError "github.com/naver/lobster/pkg/lobster/server/errors"
	"github.com/naver/lobster/pkg/lobster/util"
	"github.com/pkg/errors"
)

const LastPageNum = -1

type pageBucket struct {
	Start time.Time
	End   time.Time
	Size  uint64
}

func MakeSubQuery(req Request, seriesData model.SeriesData, pageBurst int64) (Request, model.PageInfo, uint64, error) {
	pageBuckets := makePageBuckets(req, seriesData, pageBurst)
	pageInfo := model.PageInfo{}

	pageInfo.Current = req.Page
	pageInfo.Total = len(pageBuckets)

	if req.Page == 0 || len(pageBuckets) < req.Page || req.Page < LastPageNum {
		return req, pageInfo, 0, errors.Wrap(svrError.ErrBadRequest, fmt.Sprintf("invalid page number: %d/%d page", req.Page, pageInfo.Total))
	}

	if pageInfo.Current == LastPageNum {
		pageInfo.Current = pageInfo.Total
	}

	pageInfo.HasNext = pageInfo.Current < pageInfo.Total

	subReq := req
	subReq.Version = req.Version
	subReq.Start = util.Timestamp{Time: pageBuckets[pageInfo.Current-1].Start}
	subReq.End = util.Timestamp{Time: pageBuckets[pageInfo.Current-1].End}

	return subReq, pageInfo, pageBuckets[pageInfo.Current-1].Size, nil
}

func makePageBuckets(req Request, seriesData model.SeriesData, pageBurst int64) []pageBucket {
	pageBuckets := []pageBucket{{
		Start: req.Start.Time,
		End:   req.End.Time,
	}}
	samples := seriesData.MergedSamples()
	lines := int64(0)
	size := uint64(0)

	for i, sample := range samples {
		if sample.Timestamp.Before(req.Start.Time) {
			continue
		}

		if i == 0 && req.Start.Time.Truncate(model.BucketPrecision).Before(sample.Timestamp) {
			continue
		}

		if pageBurst < lines {
			lines = sample.Lines
			size = sample.Size

			// update last page
			pageBuckets[len(pageBuckets)-1].End = sample.Timestamp

			// add next page
			pageBuckets = append(pageBuckets, pageBucket{
				Start: sample.Timestamp,
				End:   req.End.Time,
				Size:  size,
			})
		}

		lines = lines + sample.Lines
		size = size + sample.Size
	}

	return pageBuckets
}
