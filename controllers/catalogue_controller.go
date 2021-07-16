/*
Copyright 2021 pilillo.

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
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	datamillcloudv1alpha1 "github.com/pilillo/mastro-operator/api/v1alpha1"
)

// CatalogueReconciler reconciles a Catalogue object
type CatalogueReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=data-mill.cloud,resources=catalogues,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=data-mill.cloud,resources=catalogues/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=data-mill.cloud,resources=catalogues/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Catalogue object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *CatalogueReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	// retrieve catalogue instance
	catalogue := &datamillcloudv1alpha1.Catalogue{}
	err := r.Get(ctx, req.NamespacedName, catalogue)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			log.Info("Catalogue resource not found. Ignoring since object must be deleted")
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Catalogue")
		return ctrl.Result{}, err
	}

	// if catalogue resource was found
	// check if a deployment exists for the catalogue resource
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: catalogue.Name, Namespace: catalogue.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForCatalogue(catalogue)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	size := catalogue.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Ask to requeue after 1 minute in order to give enough time for the
		// pods be created on the cluster side and the operand be able
		// to do the next update step accurately.
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// Update the Memcached status with the pod names
	// List the pods for this memcached's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(catalogue.Namespace),
		client.MatchingLabels(labelsForCatalogue(catalogue.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "Catalogue.Namespace", catalogue.Namespace, "Catalogue.Name", catalogue.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, catalogue.Status.Nodes) {
		catalogue.Status.Nodes = podNames
		err := r.Status().Update(ctx, catalogue)
		if err != nil {
			log.Error(err, "Failed to update Catalogue status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CatalogueReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datamillcloudv1alpha1.Catalogue{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

// labelsForCatalogue returns the labels for selecting the resources
// belonging to the given catalogue CR name.
func labelsForCatalogue(name string) map[string]string {
	return map[string]string{"app": "mastro-catalogue", "catalogue_cr": m.Name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// deploymentForCatalogue returns a catalogue Deployment object
func (r *CatalogueReconciler) deploymentForCatalogue(m *datamillcloudv1alpha1.Catalogue) *appsv1.Deployment {
	ls := labelsForCatalogue(m.Name)

	// catalogue specific fields
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "pilillo/mastro-catalogue:20210306-static",
						Name:  "mastro-catalogue",
						//Command: []string{},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8085,
							//Name:          "catalogue",
							Protocol: "TCP",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: "/conf",
							Name:      "catalogue-conf-volume",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "catalogue-volume",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								DefaultMode: pointer.Int32Ptr(420),
							},
						},
					}},
				},
			},
		},
	}
	// Set catalogue instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}
