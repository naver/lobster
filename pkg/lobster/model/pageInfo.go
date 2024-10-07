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

package model

import "math"

// Page struct
// @Description Page inforamtion.
type PageInfo struct {
	HasNext           bool `json:"hasNext"`
	Total             int  `json:"total"`
	Current           int  `json:"current"`
	IsPartialContents bool `json:"isPartialContents"` // partial logs are returned
}

func NewPageInfo(currentPage int, lines, pageBurst int64, isPartialContents bool) PageInfo {
	totalPage := int(math.Ceil(float64(lines) / float64(pageBurst)))

	return PageInfo{
		Current:           currentPage,
		Total:             totalPage,
		HasNext:           currentPage < totalPage,
		IsPartialContents: isPartialContents,
	}
}
