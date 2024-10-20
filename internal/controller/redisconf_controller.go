/*
Copyright 2024 Seyed Javad Hosseini.

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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/javadh75/redis-k8s-operator.git/api/v1alpha1"
)

// RedisConfReconciler reconciles a RedisConf object
type RedisConfReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=api.redis.sjhr.ir,resources=redisconfs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=api.redis.sjhr.ir,resources=redisconfs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=api.redis.sjhr.ir,resources=redisconfs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RedisConf object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *RedisConfReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Start reconcile")
	// TODO(user): your logic here
	record := &apiv1alpha1.RedisConf{}
	if err := r.Get(ctx, req.NamespacedName, record); err != nil {
		if errors.IsNotFound(err) {
			pod := &corev1.Pod{}
			err = r.Get(ctx, types.NamespacedName{Name: "redis-" + req.Name, Namespace: req.Namespace}, pod)
			if err == nil {
				logger.Info("Deleting Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
				err = r.Delete(ctx, pod)
				if err != nil {
					return ctrl.Result{}, err
				}
			} else if !errors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "redis-" + record.Name,
			Namespace: req.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "redis",
					Image: "redis:" + record.Spec.ImageTag,
					Ports: []corev1.ContainerPort{
						{
							Name:          "redis",
							ContainerPort: 6379,
							Protocol:      corev1.ProtocolTCP,
						},
					},
				},
			},
		},
	}

	existingPod := &corev1.Pod{}

	if err := r.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, existingPod); err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.Create(ctx, pod)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisConfReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.RedisConf{}).
		Complete(r)
}
