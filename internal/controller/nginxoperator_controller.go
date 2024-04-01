/*
Copyright 2024 Paweł Bek.

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
	"github.com/blind3dd/nginx-operator/assets"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"path/filepath"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha1 "github.com/blind3dd/nginx-operator/api/v1alpha1"
)

// NginxOperatorReconciler reconciles a NginxOperator object
type NginxOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=operator.cloudops.com,resources=nginxoperators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.cloudops.com,resources=nginxoperators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.cloudops.com,resources=nginxoperators/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NginxOperator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *NginxOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	operatorCR := &operatorv1alpha1.NginxOperator{}
	if err := r.Get(ctx, req.NamespacedName, operatorCR); err != nil && errors.IsNotFound(err) {
		logger.Info("operator resource object not found, result: %s", ctrl.Result{})
		return ctrl.Result{}, nil
	} else if err != nil {
		logger.Error(err, "failed to get operator resource object, result: %s", ctrl.Result{})
		return ctrl.Result{}, err
	}

	deployment := &appsv1.Deployment{}
	create = false
	err := r.Get(ctx, req.NamespacedName, deployment)
	if err != nil && errors.IsNotFound(err) {
		create = true
		logger.Info("operator resource object not found, attempting to recreate.")
		var files *[]*os.File
		err := filepath.Walk("assets", findFile("0", files))
		if err != nil {
			logger.Error(err, "i/o failure, error: %v", err)
		}
		deployment = assets.GetDeploymentFromFile("./assets/nginx_deployment.yaml")
	} else if err != nil {
		logger.Error(err, "failed to get existing nginx deployment manifest.")
		return ctrl.Result{}, err
	}

	deployment.Namespace = req.Namespace
	deployment.Name = req.Name

	if operatorCR.Spec.Replicas != nil {
		deployment.Spec.Replicas = operatorCR.Spec.Replicas
	}
	if operatorCR.Spec.Port != nil {
		deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = *operatorCR.Spec.Port
	} else {
		logger.Error(err, "port is required and has not been defined ")
	}

	err = ctrl.SetControllerReference(operatorCR, deployment, r.Scheme)
	if err != nil {
		logger.Error(err, "failed to create controller operand reference.")
		return ctrl.Result{}, err
	}

	if create {
		err = r.Create(ctx, deployment)
		logger.Error(err, "deployment has been created through operator")
	} else {
		err = r.Update(ctx, deployment)
	}

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.NginxOperator{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func findFile(fileId string, files *[]*os.File) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// TODO make sure it's properly defined
		p := regexp.MustCompile("^.*_(.yaml)$").String()
		//p.ReplaceAllStringFunc(p, func(s string) string {
		//	return p.String()
		//})
		found, _ := filepath.Match(p, fileId)

		if !info.IsDir() && found {
			f, err := os.Open(path)
			if err != nil {
				panic(err)
			}
			*files = append(*files, f)
		}
		return nil
	}
}

var create bool
