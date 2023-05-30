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

package controller

import (
	"context"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	locustiov1 "locust-operator/api/v1"
)

// LocustTestReconciler reconciles a LocustTest object
type LocustTestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=locust.io,resources=locusttests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=locust.io,resources=locusttests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=locust.io,resources=locusttests/finalizers,verbs=update

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods;services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscaler,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LocustTest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *LocustTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// TODO(user): your logic here
	// Load Our Custom Object CrobJob by Name
	var locustTest locustiov1.LocustTest
	if err := r.Get(ctx, req.NamespacedName, &locustTest); err != nil {
		log.Info("Unable to fetch Custom Resource", "Err:", err)
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Define a new Deployment object
	deployment := r.deploymentForLocust(&locustTest)
	// Set Locust locustResource as the owner and controller
	if err := ctrl.SetControllerReference(&locustTest, deployment, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Get Kube Deployment Objects
	var kdeployment appsv1.Deployment
	err := r.Get(ctx, req.NamespacedName, &kdeployment)

	if err != nil && errors.IsNotFound(err) {
		// log.Error(err, "unable to list child Deployment")
		log.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.Create(ctx, deployment)
		if err != nil {
			log.Info("Created with error")
			return ctrl.Result{}, err
		}
		// Deployment created successfully - don't requeue
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// Deployment already exists - don't requeue
	log.Info("Skip reconcile: Deployment already exists", "Deployment.Namespace", kdeployment.Namespace, "Deployment.Name", kdeployment.Name)

	// Service
	service := r.serviceForLocust(&locustTest)

	// Set Locust locustResource as the owner and controller
	if err := ctrl.SetControllerReference(&locustTest, service, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Check if this Service already exists
	foundsvc := &corev1.Service{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundsvc)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.Create(context.TODO(), service)
		if err != nil {
			return ctrl.Result{}, err
		}

		// Service created successfully - don't requeue
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// Service already exists - don't requeue
	log.Info("Skip reconcile: Service already exists", "Service.Namespace", foundsvc.Namespace, "Service.Name", foundsvc.Name)

	// Locust worker deployment, limit for maximum number of slaves set to 30
	if locustTest.Spec.NumWorkers != 0 && locustTest.Spec.NumWorkers < 30 {
		slavedeployment := r.deploymentForLocustSlaves(&locustTest)

		// Set Locust locustResource as the owner and controller
		if err := ctrl.SetControllerReference(&locustTest, slavedeployment, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		// Check if this Deployment already exists
		foundslaves := &appsv1.Deployment{}
		err = r.Get(context.TODO(), types.NamespacedName{Name: slavedeployment.Name, Namespace: slavedeployment.Namespace}, foundslaves)
		if err != nil && errors.IsNotFound(err) {
			log.Info("Creating a new Locust Worker Deployment", "Deployment.Namespace", slavedeployment.Namespace, "Deployment.Name", slavedeployment.Name)
			err = r.Create(context.TODO(), slavedeployment)
			if err != nil {
				return ctrl.Result{}, err
			}

			// Deployment created successfully - don't requeue
			return ctrl.Result{}, nil
		} else if err != nil {
			return ctrl.Result{}, err
		}

		// Deployment already exists - don't requeue
		log.Info("Skip reconcile: Locust Worker Deployment already exists", "Deployment.Namespace", foundslaves.Namespace, "Deployment.Name", foundslaves.Name)

	}

	// List Kube Pods
	var podList corev1.PodList
	if err := r.List(ctx, &podList, client.InNamespace(req.Namespace)); err != nil {
		log.Error(err, "Unable to list child Pods!")
		return ctrl.Result{}, err
	}

	// Loop over Deployment and Pod Lists
	deploymentFound := false
	// podsFound := false
	// if kdeployment.GetName() == locustTest.Name {
	// 	log.Info("deployment linked to a custom resource found", "name", deployment.GetName())
	// 	deploymentFound = true

	// 	if len(podList.Items) == int(locustTest.Spec.NumPods) {
	// 		log.Info("Number of Pods linked to a custom resource matched", "numPods", len(podList.Items))
	// 		podsFound = true
	// 	} else {
	// 		log.Info("Insufficient Pods! Waiting for Kube Deployment Controller to scale.")
	// 	}
	// } else {
	// 	log.Info("Deployment does not match CR!", deployment.GetName(), locustTest.Name)
	// }

	// Update Status
	locustTest.Status.DeploymentHappy = deploymentFound
	// locustTest.Status.PodsHappy = podsFound
	if err := r.Status().Update(ctx, &locustTest); err != nil {
		log.Error(err, "Unable to update locustTest status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LocustTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&locustiov1.LocustTest{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

// deploymentForLocust returns a Locust Deployment object
func (r *LocustTestReconciler) deploymentForLocust(cr *locustiov1.LocustTest) *appsv1.Deployment {
	type label map[string]string
	ls := label{"app": "Locust"}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: cr.Spec.Image,
						Name:  cr.Name,
						// TODO: refactor into function
						Command: []string{"locust", "-f", "/mnt/locust/main.py", "--master", "-H", cr.Spec.HostUrl, "-u", strconv.Itoa(int(cr.Spec.Users)), "-r", strconv.Itoa(int(cr.Spec.SpawnRate)), "--run-time", cr.Spec.RunTime, "--autostart"},
						Env: []corev1.EnvVar{
							{
								Name:  "TARGET_HOST",
								Value: cr.Spec.HostUrl,
							},
						},
						Ports: []corev1.ContainerPort{
							{
								Name:          "http",
								Protocol:      corev1.ProtocolTCP,
								ContainerPort: 8089,
							},
							{
								Name:          "worker-1",
								Protocol:      corev1.ProtocolTCP,
								ContainerPort: 5557,
							},
							{
								Name:          "worker-2",
								Protocol:      corev1.ProtocolTCP,
								ContainerPort: 5558,
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "locustfile",
								MountPath: "/mnt/locust",
							},
							{
								Name:      "locustsecret",
								MountPath: cr.Spec.SecretPath,
							},
						},
					}},
					Volumes: []corev1.Volume{
						{
							Name: "locustfile",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Spec.ConfigMap,
									},
								}},
						},
						{
							Name: "locustsecret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: cr.Spec.Secret,
								}},
						},
					},
				},
			},
		},
	}
	// Set Locust locustResource as the owner and controller
	ctrl.SetControllerReference(cr, dep, r.Scheme)
	return dep
}

// deploymentForLocustSlaves returns a Locust Deployment object
func (r *LocustTestReconciler) deploymentForLocustSlaves(cr *locustiov1.LocustTest) *appsv1.Deployment {
	type label map[string]string
	ls := label{"app": "LocustWorkers"}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-worker",
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &cr.Spec.NumWorkers,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   cr.Spec.Image,
						Name:    cr.Name + "-worker",
						Command: []string{"locust", "--worker", "--master-host", cr.Name + "-service", "-f", "/mnt/locust/main.py"},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "locustfile",
								MountPath: "/mnt/locust",
							},
							{
								Name:      "locustsecret",
								MountPath: cr.Spec.SecretPath,
							},
						},
					}},
					Volumes: []corev1.Volume{
						{
							Name: "locustfile",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Spec.ConfigMap,
									},
								}},
						},
						{
							Name: "locustsecret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: cr.Spec.Secret,
								}},
						},
					},
				},
			},
		},
	}
	// Set Locust locustResource as the owner and controller
	ctrl.SetControllerReference(cr, dep, r.Scheme)
	return dep
}

// serviceForLocust returns a Service object
func (r *LocustTestReconciler) serviceForLocust(cr *locustiov1.LocustTest) *corev1.Service {
	type label map[string]string
	ls := label{"app": "Locust"}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-service",
			Namespace: cr.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Protocol: corev1.ProtocolTCP,
					Port:     8089,
				},
				{
					Name:     "worker-1",
					Protocol: corev1.ProtocolTCP,
					Port:     5557,
				},
				{
					Name:     "worker-2",
					Protocol: corev1.ProtocolTCP,
					Port:     5558,
				},
			},
		},
	}
	// Set Locust locustResource as the owner and controller
	ctrl.SetControllerReference(cr, svc, r.Scheme)
	return svc
}

func int32Ptr(i int32) *int32 { return &i }
