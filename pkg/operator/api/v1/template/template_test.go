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

package template

import (
	"testing"
	"time"
)

func TestGeneratePath(t *testing.T) {
	templates := map[string]string{
		"/my":                       "/my",
		"/{{TimeFunc \"2006-01\"}}": "/2025-01",
		"/{{.Pod}}":                 "/loggen-pod",
		"/{{.Container}}/{{.Pod}}":  "/container1/loggen-pod",
		"/{{.SourcePath}}":          "/renamed_namespaceA_test.log",
		"/my/{{TimeFunc \"2006-01\"}}/123/{{.SourcePath}}":                                     "/my/2025-01/123/renamed_namespaceA_test.log",
		"/{{.Pod}}-{{.Container}}/{{.Pod}}":                                                    "/loggen-pod-container1/loggen-pod",
		"/{{TimeFunc \"2006-01\"}}/{{TimeFunc \"2006-01-02\"}}/{{TimeFunc \"2006-01-02_15\"}}": "/2025-01/2025-01-06/2025-01-06_14",
	}

	timeInput := time.Date(2025, 1, 6, 14, 17, 15, 0, time.UTC)

	data := PathElement{
		Namespace:  "namespaceA",
		SinkName:   "sink1",
		Pod:        "loggen-pod",
		Container:  "container1",
		SourcePath: "renamed_namespaceA_test.log",
		TimeInput:  timeInput,
	}

	for tmpl, expected := range templates {
		path, err := GeneratePath(tmpl, data)
		if err != nil {
			t.Errorf("failed to generating path for template %q: %v\n", tmpl, err)
			return
		}
		if path != expected {
			t.Errorf("invalid result: %s vs %s", path, expected)
			return
		}

		t.Logf("template: %q\npath: %s", tmpl, path)
	}
}

func TestInvalidPath(t *testing.T) {
	templates := []string{
		"/{.SourcePath}}",            // invalid brace
		"/{{.TimeFunc \"2006-01\"}}", // invalid function usage
		"/{{TimeFunc}}",              // invalid function usage
	}

	for _, tmpl := range templates {
		err := ValidateTemplateString(tmpl)
		if err == nil {
			t.Errorf("expected an error but got none: validation failed for %s", tmpl)
		}
	}
}
