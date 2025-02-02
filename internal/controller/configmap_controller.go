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
	"log/slog"
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	opsappv1 "someapp.cn/configmap-restart/api/v1"
)

// ConfigMapReconciler reconciles a ConfigMapReconciler object
type ConfigMapReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *ConfigMapReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("configmap-controller", req.NamespacedName)
	log.Info("Reconciling configmap")

	configMap := &corev1.ConfigMap{}
	result := ctrl.Result{}

	err := r.Get(ctx, req.NamespacedName, configMap)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			// configMap delete, do nothing
			return result, nil
		}
		log.Error(err, "failed to get configMap")
		return ctrl.Result{Requeue: false}, client.IgnoreNotFound(err)
	}

	// main logic
	configRestartList := &opsappv1.ConfigrestartList{}
	err = r.List(ctx, configRestartList, &client.ListOptions{
		Namespace: req.Namespace,
	})
	if err != nil {
		log.Error(err, "failed to list configrestart")
		return ctrl.Result{Requeue: false}, client.IgnoreNotFound(err)
	}

	for _, configRestart := range configRestartList.Items {

		go func(configRestart *opsappv1.Configrestart) {
			// go func(deployments []string, configName, namespace string, suspend bool) {
			suspend := configRestart.Spec.Suspend
			configMapName := configRestart.Spec.ConfigName
			deploymentNames := configRestart.Spec.Deployments
			ns := configRestart.Namespace

			if configMapName != configMap.Name {
				log.Info("configName not match, skip restart", "configMapName", configMap.Name)
				return
			}

			if suspend {
				log.Info("configrestart is suspended, skip restart", "configName", configMapName)
				return
			}
			// if deploymentName is empty, restart all deployments
			if len(deploymentNames) == 0 {

				err := restartDeploymentWithConfigMap(ctx, r.Client, configMapName, ns)
				if err != nil {
					log.Error(err, "restart deployment with configmap failed")
					return
				}
				log.Info("restart deployment with configmap success", "configName", configMapName)
				return
			}
			// if deployments is not empty, restart deployments
			for _, deploymentName := range deploymentNames {
				err := restartDeployment(ctx, r.Client, deploymentName, ns)
				if err != nil {
					log.Error(err, "restart deployment failed", "deployment", deploymentName, "ns", ns)
					return
				}
				log.Info("restart deployment success", "deployment", deploymentName, "ns", ns)
			}
		}(&configRestart)
	}

	log.Info("Reconciling configmap finished")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigMapReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}, builder.WithPredicates(configMapUpdatePredicate())).
		Complete(r)
}

func restartDeployment(ctx context.Context, c client.Client, deploymentName, ns string) error {

	var deploy appsv1.Deployment

	err := c.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: ns}, &deploy)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			slog.Info("Deployment not found, skip restart", "deployment", deploymentName, "namespace", ns)
			return nil
		}
		return err
	}

	if deploy.Spec.Template.Annotations == nil {
		deploy.Spec.Template.Annotations = make(map[string]string)
	}
	deploy.Spec.Template.Annotations["configrestart/restart"] = time.Now().Format(time.RFC3339)

	err = c.Update(ctx, &deploy, &client.UpdateOptions{FieldManager: "configrestart-controller"})
	if err != nil {
		return err
	}

	return nil
}
func checkDeploymentHasConfigMap(configMapName string, deploy *appsv1.Deployment) bool {

	for _, volume := range deploy.Spec.Template.Spec.Volumes {
		if volume.ConfigMap != nil && volume.ConfigMap.Name == configMapName {
			return true
		}
	}
	return false
}

func restartDeploymentWithConfigMap(ctx context.Context, c client.Client, configName, ns string) error {

	var deploys appsv1.DeploymentList
	err := c.List(ctx, &deploys, &client.ListOptions{Namespace: ns})
	if err != nil {
		return err
	}

	for _, deploy := range deploys.Items {
		if checkDeploymentHasConfigMap(configName, &deploy) {
			err := restartDeployment(ctx, c, deploy.Name, ns)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func configMapUpdatePredicate() predicate.Funcs {
	controllerLeader := "control-plane.alpha.kubernetes.io/leader"
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldCM, okOld := e.ObjectOld.(*corev1.ConfigMap)
			newCM, okNew := e.ObjectNew.(*corev1.ConfigMap)
			if !okOld || !okNew {
				return false
			}
			//
			if newCM.GetAnnotations()[controllerLeader] != "" {
				return false
			}
			return !reflect.DeepEqual(oldCM.Data, newCM.Data)
		},
		CreateFunc:  func(e event.CreateEvent) bool { return false },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}
}
