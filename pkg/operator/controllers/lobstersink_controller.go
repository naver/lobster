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

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sinkV1 "github.com/naver/lobster/pkg/operator/api/v1"
)

// LobsterSinkReconciler reconciles a LobsterSink object.
type LobsterSinkReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=lobster.io,resources=lobstersinks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=lobster.io,resources=lobstersinks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=lobster.io,resources=lobstersinks/finalizers,verbs=update

// SetupWithManager sets up the controller with the Manager.
func (r *LobsterSinkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sinkV1.LobsterSink{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LobsterSink object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *LobsterSinkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	instance := &sinkV1.LobsterSink{}
	if err := r.Get(context.TODO(), req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// TODO: inspection

	if instance.Status.Init != sinkV1.StatusInitSucceeded {
		return ctrl.Result{}, r.updateStatus(instance, sinkV1.StatusInitSucceeded)
	}

	return ctrl.Result{}, nil
}

func (r *LobsterSinkReconciler) updateStatus(instance *sinkV1.LobsterSink, statusInit string) error {
	if instance.Status.Init == statusInit {
		return nil
	}

	instance.Status.Init = statusInit
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.Status().Update(context.TODO(), instance)
	})
}
