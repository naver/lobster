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
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	cache "github.com/hashicorp/golang-lru"
)

var templateCache *cache.Cache

func init() {
	templateCache, _ = cache.New(200)
}

type PathElement struct {
	Cluster    string
	Namespace  string
	SinkName   string
	RuleName   string
	Pod        string
	Container  string
	SourceType string
	SourcePath string
	TimeInput  time.Time
}

func (d PathElement) TimeFormat(layout string) string {
	return d.TimeInput.Format(layout)
}

func ValidateTemplateString(templateStr string) error {
	if len(templateStr) == 0 {
		return nil
	}

	if strings.Count(templateStr, "{{") != strings.Count(templateStr, "}}") {
		return errors.New("mismatch between '{{' and '}}'")
	}

	if !filepath.IsAbs(templateStr) {
		return errors.New("the template should be an absolute path (starting with `/`)")
	}

	tmpl, err := getTemplate(fmt.Sprintf("validate_%s", templateStr), PathElement{}).Parse(templateStr)
	if err != nil {
		return err
	}

	return tmpl.Execute(&bytes.Buffer{}, PathElement{})
}

func GeneratePath(templateStr string, elem PathElement) (string, error) {
	var result bytes.Buffer

	tmpl, err := getTemplate(templateStr, elem).Parse(templateStr)
	if err != nil {
		return "", err
	}

	if err := tmpl.Execute(&result, elem); err != nil {
		return "", err
	}

	path, err := url.PathUnescape(url.PathEscape(result.String()))
	if err != nil {
		return "", err
	}

	return path, nil
}

func getTemplate(templateStr string, elem PathElement) *template.Template {
	v, ok := templateCache.Get(templateStr)
	if !ok {
		newTmpl := template.New(templateStr).Funcs(template.FuncMap{
			"TimeLayout": elem.TimeFormat,
		})
		templateCache.Add(templateStr, newTmpl)

		return newTmpl
	}

	return v.(*template.Template)
}
