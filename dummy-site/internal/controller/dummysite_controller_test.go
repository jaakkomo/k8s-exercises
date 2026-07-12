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

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kapps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	webv1 "github.com/jaakkomo/k8s-exercises/dummy-site/api/v1"
)

var _ = Describe("DummySite Controller", func() {
	Context("When reconciling a resource", func() {
		var (
			resourceName       string
			resourceNamespace  = "default"
			typeNamespacedName types.NamespacedName
			ctx                = context.Background()
			dummysite          = &webv1.DummySite{}
		)

		BeforeEach(func() {
			By("creating the custom resource for the Kind DummySite")

			resourceName = "test-" + uuid.NewString()
			typeNamespacedName = types.NamespacedName{
				Name:      resourceName,
				Namespace: resourceNamespace,
			}

			dummysite = &webv1.DummySite{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: resourceNamespace,
				},
				Spec: webv1.DummySiteSpec{
					WebsiteURL:       "https://example.com",
					GatewayName:      "test-gateway",
					GatewayNamespace: "test-gateway-namespace",
					Hostname:         "testing.localhost",
				},
			}

			Expect(k8sClient.Create(ctx, dummysite)).To(Succeed())
			Expect(k8sClient.Get(ctx, typeNamespacedName, dummysite)).To(Succeed())
		})

		AfterEach(func() {
			resource := &webv1.DummySite{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance DummySite")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should create a Deployment, Service and HTTPRoute", func() {
			By("Reconciling the created resource")
			controllerReconciler := &DummySiteReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the Deployment")
			var deployment kapps.Deployment
			Expect(k8sClient.Get(ctx, typeNamespacedName, &deployment)).To(Succeed())
			Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(Equal("nginx:1.31.2-alpine"))
			Expect(deployment.Spec.Template.Spec.InitContainers).To(HaveLen(1))
			Expect(deployment.Spec.Template.Spec.InitContainers[0].Env[0].Value).To(Equal("https://example.com"))

			By("Checking the Service")
			var service core.Service
			Expect(k8sClient.Get(ctx, typeNamespacedName, &service)).To(Succeed())
			Expect(service.Spec.Selector).To(HaveKeyWithValue("app", resourceName))
			Expect(service.Spec.Ports).To(HaveLen(1))
			Expect(service.Spec.Ports[0].Port).To(Equal(int32(80)))

			By("Checking the HTTPRoute")
			var route gatewayv1.HTTPRoute
			Expect(k8sClient.Get(ctx, typeNamespacedName, &route)).To(Succeed())
			Expect(route.Spec.ParentRefs).To(HaveLen(1))
			Expect(route.Spec.CommonRouteSpec.ParentRefs).To(HaveLen(1))
			Expect(route.Spec.CommonRouteSpec.ParentRefs[0].Name).To(Equal(gatewayv1.ObjectName("test-gateway")))
			Expect(string(*route.Spec.CommonRouteSpec.ParentRefs[0].Namespace)).To(Equal("test-gateway-namespace"))
			Expect(route.Spec.Hostnames).To(HaveLen(1))
			Expect(route.Spec.Hostnames[0]).To(Equal(gatewayv1.Hostname("testing.localhost")))
			Expect(route.Spec.Rules).To(HaveLen(1))
			Expect(route.Spec.Rules[0].BackendRefs).To(HaveLen(1))
			Expect(route.Spec.Rules[0].BackendRefs[0].Name).To(Equal(gatewayv1.ObjectName(resourceName)))
		})

		It("should set the DummySite as owner of all created resources", func() {
			controllerReconciler := &DummySiteReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			var deployment kapps.Deployment
			Expect(k8sClient.Get(ctx, typeNamespacedName, &deployment)).To(Succeed())
			Expect(metav1.IsControlledBy(&deployment, dummysite)).To(BeTrue())

			var service core.Service
			Expect(k8sClient.Get(ctx, typeNamespacedName, &service)).To(Succeed())
			Expect(metav1.IsControlledBy(&service, dummysite)).To(BeTrue())

			var route gatewayv1.HTTPRoute
			Expect(k8sClient.Get(ctx, typeNamespacedName, &route)).To(Succeed())
			Expect(metav1.IsControlledBy(&route, dummysite)).To(BeTrue())
		})
	})
})
