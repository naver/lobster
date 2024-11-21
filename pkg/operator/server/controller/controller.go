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

package controller

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"
	sinkV1 "github.com/naver/lobster/pkg/operator/api/v1"
	v1 "github.com/naver/lobster/pkg/operator/server/api/v1"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ErrImproperParam       = fmt.Errorf("improper parameters")
	ErrUnsupportedType     = fmt.Errorf("unsupported sink type")
	ErrNotFound            = fmt.Errorf("resource is not found")
	ErrUnprocessableEntity = fmt.Errorf("may not create more resources due to limit")

	defaultTimeout = 5 * time.Second
)

type SinkController struct {
	Client     client.Client
	MaxContent int
	Logger     logr.Logger
}

func (c SinkController) List(namespace, name, sinkType string) ([]v1.Sink, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), defaultTimeout)
	defer cancel()

	sinks := []v1.Sink{}
	result := &sinkV1.LobsterSinkList{}

	if err := c.Client.List(ctx, result, &client.ListOptions{Namespace: namespace}); err != nil {
		if errors.IsNotFound(err) {
			return sinks, nil
		}
		return sinks, err
	}

	for _, item := range result.Items {
		if len(sinkType) > 0 && item.Spec.SinkType != sinkType {
			continue
		}

		if len(name) > 0 && len(name) != 0 && item.Name != name {
			continue
		}

		sinks = append(sinks, v1.Sink{
			Name:           item.Name,
			Namespace:      item.Namespace,
			Type:           item.Spec.SinkType,
			Description:    item.Spec.Description,
			LogMetricRules: item.Spec.LogMetricRules,
			LogExportRules: item.Spec.LogExportRules,
		})
	}

	return sinks, nil
}

func (c SinkController) Put(sink v1.Sink) (bool, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), defaultTimeout)
	defer cancel()

	result := &sinkV1.LobsterSink{}
	if err := c.Client.Get(ctx, types.NamespacedName{
		Namespace: sink.Namespace,
		Name:      sink.Name,
	}, result); err != nil {
		if !errors.IsNotFound(err) {
			return false, err
		}

		return true, retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			return c.Client.Create(context.TODO(), &sinkV1.LobsterSink{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sink.Name,
					Namespace: sink.Namespace,
				},
				Spec: sinkV1.LobsterSinkSpec{
					SinkType:       sink.Type,
					LogMetricRules: sink.LogMetricRules,
					LogExportRules: sink.LogExportRules,
					Description:    sink.Description,
				},
			})
		})
	}

	if result.Spec.SinkType != sink.Type {
		return false, ErrNotFound
	}

	switch result.Spec.SinkType {
	case sinkV1.LogMetricRules:
		rules := v1.MergeContent(result.Spec.LogMetricRules, sink.LogMetricRules).([]sinkV1.LogMetricRule)
		if c.MaxContent < len(rules) {
			return false, ErrUnprocessableEntity
		}
		result.Spec.LogMetricRules = rules
	case sinkV1.LogExportRules:
		rules := v1.MergeContent(result.Spec.LogExportRules, sink.LogExportRules).([]sinkV1.LogExportRule)
		if c.MaxContent < len(rules) {
			return false, ErrUnprocessableEntity
		}
		result.Spec.LogExportRules = rules
	}

	if len(sink.Description) != 0 {
		result.Spec.Description = sink.Description
	}

	return false, retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return c.Client.Update(context.TODO(), result)
	})
}

func (c SinkController) Delete(namespace, name string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), defaultTimeout)
	defer cancel()

	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		if err := c.Client.Delete(ctx, &sinkV1.LobsterSink{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      name,
			},
		}); err != nil {
			if errors.IsNotFound(err) {
				return ErrNotFound
			}

			return err
		}

		return nil
	})
}

func (c SinkController) DeleteContent(namespace, name, ruleName string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), defaultTimeout)
	defer cancel()

	sink := &sinkV1.LobsterSink{}
	if err := c.Client.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, sink); err != nil {
		if errors.IsNotFound(err) {
			return ErrNotFound
		}

		return err
	}

	switch sink.Spec.SinkType {
	case sinkV1.LogMetricRules:
		index := v1.SearchContentToDelete(sink.Spec.LogMetricRules, ruleName)
		if index < 0 {
			return ErrNotFound
		}
		sink.Spec.LogMetricRules = append(sink.Spec.LogMetricRules[:index], sink.Spec.LogMetricRules[index+1:]...)

		if len(sink.Spec.LogMetricRules) == 0 {
			return c.Delete(namespace, name)
		}
	case sinkV1.LogExportRules:
		index := v1.SearchContentToDelete(sink.Spec.LogExportRules, ruleName)
		if index < 0 {
			return ErrNotFound
		}
		sink.Spec.LogExportRules = append(sink.Spec.LogExportRules[:index], sink.Spec.LogExportRules[index+1:]...)

		if len(sink.Spec.LogExportRules) == 0 {
			return c.Delete(namespace, name)
		}
	default:
		return ErrUnsupportedType
	}

	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return c.Client.Update(ctx, sink)
	})
}
