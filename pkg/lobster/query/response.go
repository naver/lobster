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

import "github.com/naver/lobster/pkg/lobster/model"

// Response struct
// @Description Response wrapping series and logs from store.
type Response struct {
	SeriesData *model.SeriesData `json:"series,omitempty"`   // timeseries data
	Contents   string            `json:"contents"`           // logs in string
	PageInfo   *model.PageInfo   `json:"pageInfo,omitempty"` // page information
}

// ResponseEntries struct
// @Description Response wrapping series and logs from querier.
type ResponseEntries struct {
	SeriesData *model.SeriesData `json:"series,omitempty"`   // timeseries data
	Contents   []model.Entry     `json:"contents"`           // log entries
	PageInfo   *model.PageInfo   `json:"pageInfo,omitempty"` // page information
}
