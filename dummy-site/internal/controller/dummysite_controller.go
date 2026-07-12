/*
Copyright 2026.

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

	kapps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	webv1 "github.com/jaakkomo/k8s-exercises/dummy-site/api/v1"
)

// DummySiteReconciler reconciles a DummySite object
type DummySiteReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func constructDeploymentForDummySite(dummySite *webv1.DummySite, r *DummySiteReconciler) (*kapps.Deployment, error) {
	deployment := &kapps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dummySite.Name,
			Namespace: dummySite.Namespace,
		},
		Spec: kapps.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": dummySite.Name,
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": dummySite.Name,
					},
				},
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						{
							Name: "website",
							VolumeSource: core.VolumeSource{
								EmptyDir: &core.EmptyDirVolumeSource{},
							},
						},
					},
					Containers: []core.Container{
						{
							Name:  "nginx",
							Image: "nginx:1.31.2-alpine",
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "website",
									MountPath: "/usr/share/nginx/html",
								},
							},
						},
					},
					InitContainers: []core.Container{
						{
							Name:  "init-download-website",
							Image: "curlimages/curl:8.21.0",
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "website",
									MountPath: "/website",
								},
							},
							Env: []core.EnvVar{
								{
									Name:  "WEBSITE_URL",
									Value: dummySite.Spec.WebsiteURL,
								},
							},
							Command: []string{
								"sh",
								"-c",
								"curl -L \"$WEBSITE_URL\" -o /website/index.html",
							},
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(dummySite, deployment, r.Scheme); err != nil {
		return nil, err
	}

	return deployment, nil
}

func constructServiceForDummySite(dummySite *webv1.DummySite, r *DummySiteReconciler) (*core.Service, error) {
	service := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dummySite.Name,
			Namespace: dummySite.Namespace,
		},
		Spec: core.ServiceSpec{
			Type: core.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app": dummySite.Name,
			},
			Ports: []core.ServicePort{
				{
					Name:     "http",
					Port:     80,
					Protocol: core.ProtocolTCP,
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(dummySite, service, r.Scheme); err != nil {
		return nil, err
	}

	return service, nil
}

func constructHTTPRouteForDummySite(dummySite *webv1.DummySite, r *DummySiteReconciler) (*gatewayv1.HTTPRoute, error) {
	httpRoute := &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dummySite.Name,
			Namespace: dummySite.Namespace,
		},
		Spec: gatewayv1.HTTPRouteSpec{
			CommonRouteSpec: gatewayv1.CommonRouteSpec{
				ParentRefs: []gatewayv1.ParentReference{
					{
						Name: gatewayv1.ObjectName(dummySite.Spec.GatewayName),
						Namespace: func() *gatewayv1.Namespace {
							ns := gatewayv1.Namespace(dummySite.Spec.GatewayNamespace)
							return &ns
						}(),
					},
				},
			},
			Hostnames: []gatewayv1.Hostname{
				gatewayv1.Hostname(dummySite.Spec.Hostname),
			},
			Rules: []gatewayv1.HTTPRouteRule{
				{
					BackendRefs: []gatewayv1.HTTPBackendRef{
						{
							BackendRef: gatewayv1.BackendRef{
								BackendObjectReference: gatewayv1.BackendObjectReference{
									Name: gatewayv1.ObjectName(dummySite.Name),
									Port: func() *gatewayv1.PortNumber {
										p := gatewayv1.PortNumber(80)
										return &p
									}(),
								},
							},
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(dummySite, httpRoute, r.Scheme); err != nil {
		return nil, err
	}

	return httpRoute, nil
}

func (r *DummySiteReconciler) ensureResource(
	ctx context.Context,
	key client.ObjectKey,
	resource client.Object,
	construct func() (client.Object, error),
) error {
	err := r.Get(ctx, key, resource)
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	if apierrors.IsNotFound(err) {
		resource, err := construct()
		if err != nil {
			return err
		}
		return r.Create(ctx, resource)
	}
	return nil
}

// +kubebuilder:rbac:groups=web.jaakkomo.dwk,resources=dummysites,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=web.jaakkomo.dwk,resources=dummysites/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=web.jaakkomo.dwk,resources=dummysites/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=httproutes,verbs=get;list;watch;create;update;patch

func (r *DummySiteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var dummySite webv1.DummySite
	if err := r.Get(ctx, req.NamespacedName, &dummySite); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.ensureResource(
		ctx,
		req.NamespacedName,
		&kapps.Deployment{},
		func() (client.Object, error) {
			log.Info("Creating Deployment")
			return constructDeploymentForDummySite(&dummySite, r)
		},
	); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.ensureResource(
		ctx,
		req.NamespacedName,
		&core.Service{},
		func() (client.Object, error) {
			log.Info("Creating Service")
			return constructServiceForDummySite(&dummySite, r)
		},
	); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.ensureResource(
		ctx,
		req.NamespacedName,
		&gatewayv1.HTTPRoute{},
		func() (client.Object, error) {
			log.Info("Creating HTTPRoute")
			return constructHTTPRouteForDummySite(&dummySite, r)
		},
	); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DummySiteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webv1.DummySite{}).
		Named("dummysite").
		Complete(r)
}
