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

/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"fmt"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	StatusInitInitializing = "initializing"
	StatusInitSucceeded    = "succeeded"
	StatusInitFailed       = "failed"

	LogExportRules = "logExportRules"
	LogMetricRules = "logMetricRules"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LobsterSinkSpec defines the desired state of LobsterSink.
type LobsterSinkSpec struct {
	// Type that distinguishes logMetricRules and logExportRules
	SinkType string `json:"type,omitempty"`
	// Description of this custom resource
	Description string `json:"description,omitempty"`
	// Rules for generating log metrics
	LogMetricRules []LogMetricRule `json:"logMetricRules,omitempty"`
	// Rules for exporting logs
	LogExportRules []LogExportRule `json:"logExportRules,omitempty"`
	Limit          int             `json:"limit,omitempty"`
	TimeZone       string          `json:"timezone,omitempty"`
}

type Filter struct {
	// Filter logs only for specific Namespace
	Namespace string `json:"namespace,omitempty"`
	// Filter logs only for specific Clusters
	Clusters []string `json:"clusters,omitempty"`
	// Filter logs only for specific Pod labels
	Labels []map[string]string `json:"labels,omitempty"`
	// Filter logs only for specific ReplicaSets/StatefulSets
	SetNames []string `json:"setNames,omitempty"`
	// Filter logs only for specific Pods
	Pods []string `json:"pods,omitempty"`
	// Filter logs only for specific Containers
	Containers []string `json:"containers,omitempty"`
	// Filter logs only for specific Sources
	Sources []Source `json:"sources,omitempty"`
	// Filter only logs that match the re2 expression(https://github.com/google/re2/wiki/Syntax)
	FilterIncludeExpr string `json:"include,omitempty"`
	// Filter only logs that do not match the re2 expression(https://github.com/google/re2/wiki/Syntax)
	FilterExcludeExpr string `json:"exclude,omitempty"`
}

func (f Filter) Validate() error {
	if len(f.Namespace) == 0 {
		return fmt.Errorf("`namespace` should not be empty")
	}

	if len(f.SetNames) == 0 && len(f.Pods) == 0 && len(f.Labels) == 0 {
		return fmt.Errorf("`setNames` or `pods` or `labels` should not be empty")
	}

	if len(f.FilterIncludeExpr) > 0 {
		if _, err := regexp.Compile(f.FilterIncludeExpr); err != nil {
			return err
		}
	}

	if len(f.FilterExcludeExpr) > 0 {
		if _, err := regexp.Compile(f.FilterExcludeExpr); err != nil {
			return err
		}
	}

	return nil
}

type Source struct {
	Type string `json:"type,omitempty"`
	Path string `json:"path,omitempty"`
}

// LobsterSinkStatus defines the observed state of LobsterSink.
type LobsterSinkStatus struct {
	Init string `json:"init,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LobsterSink is the Schema for the lobstersinks API.
type LobsterSink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LobsterSinkSpec   `json:"spec,omitempty"`
	Status LobsterSinkStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LobsterSinkList contains a list of LobsterSink.
type LobsterSinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LobsterSink `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LobsterSink{}, &LobsterSinkList{})
}
