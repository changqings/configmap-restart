/*
Copyright 2024.

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

package controller

import (
	"context"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	opsappv1 "someapp.cn/configmap-restart/api/v1"
)

// ConfigrestartReconciler reconciles a Configrestart object
type ConfigrestartReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=opsapp.someapp.cn,resources=configrestarts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=opsapp.someapp.cn,resources=configrestarts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=opsapp.someapp.cn,resources=configrestarts/finalizers,verbs=update

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *ConfigrestartReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("configrestart", req.NamespacedName)
	log.Info("Reconciling configrestart")

	configRestart := &opsappv1.Configrestart{}
	result := ctrl.Result{}

	// get configRestart, and write into configRestart
	err := r.Get(ctx, req.NamespacedName, configRestart)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			// not found, must be delete
			log.Info("deleted configrestart SharedMap and ShareMapIndex", "key", req.NamespacedName)
			return result, nil
		}
		log.Error(err, "failed to get configrestart")
		return ctrl.Result{Requeue: false}, client.IgnoreNotFound(err)
	}
	// we do nothing here, just for caching configRestart

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigrestartReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsappv1.Configrestart{}).
		Complete(r)
}
