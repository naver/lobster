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

package syncer

import (
	"flag"
	"time"
)

type config struct {
	LobsterSinkOperator *string
	SyncInterval        *time.Duration
}

func setup() config {
	lobsterSinkOperator := flag.String("syncer.lobsterSinkOperator", "lobster-operator:80", "host to get log metric/export info")
	syncInterval := flag.Duration("syncer.syncInterval", 30*time.Second, "sync interval")

	return config{
		LobsterSinkOperator: lobsterSinkOperator,
		SyncInterval:        syncInterval,
	}
}
