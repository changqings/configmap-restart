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
	"sync"
	"time"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	opsappv1 "someapp.cn/configmap-restart/api/v1"
)

var (
	SharedMap     = make(map[types.NamespacedName]ConfigMapData)
	ShareMapIndex = make(map[ConfigMapNamespaceName]types.NamespacedName)
	Mutex         sync.Mutex
)

type ConfigMapNamespaceName struct {
	Namespace string
	Name      string
}

type ConfigMapData struct {
	ConfigName  string
	Deployments []string
	Suspend     bool
}

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
			Mutex.Lock()
			delete(SharedMap, req.NamespacedName)
			deleteShareMapIndexKey(ShareMapIndex, req.NamespacedName)
			Mutex.Unlock()
			log.Info("deleted configrestart SharedMap and ShareMapIndex", "key", req.NamespacedName)
			return result, nil
		}
		log.Error(err, "failed to get configrestart")
		return ctrl.Result{RequeueAfter: time.Second * 10}, client.IgnoreNotFound(err)
	}
	//
	configMapData := ConfigMapData{
		ConfigName:  configRestart.Spec.ConfigName,
		Deployments: configRestart.Spec.Deployments,
		Suspend:     configRestart.Spec.Suspend,
	}
	configMapNamespaceName := ConfigMapNamespaceName{
		Namespace: configRestart.Namespace,
		Name:      configRestart.Spec.ConfigName,
	}

	Mutex.Lock()
	SharedMap[req.NamespacedName] = configMapData
	ShareMapIndex[configMapNamespaceName] = req.NamespacedName
	Mutex.Unlock()
	log.Info("stored configrestart shared map")

	return ctrl.Result{}, nil
}

func deleteShareMapIndexKey(mapIndex map[ConfigMapNamespaceName]types.NamespacedName, value types.NamespacedName) {
	for k, v := range mapIndex {
		if v == value {
			delete(ShareMapIndex, k)
			return
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigrestartReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsappv1.Configrestart{}).
		Complete(r)
}
