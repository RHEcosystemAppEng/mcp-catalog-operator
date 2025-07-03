/*
Copyright 2025.

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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
)

var _ = Describe("McpServerCertJob Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		mcpservercertjob := &mcpv1alpha1.McpServerCertJob{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind McpServerCertJob")
			err := k8sClient.Get(ctx, typeNamespacedName, mcpservercertjob)
			if err != nil && errors.IsNotFound(err) {
				resource := &mcpv1alpha1.McpServerCertJob{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					// TODO(user): Specify other spec details if needed.
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &mcpv1alpha1.McpServerCertJob{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance McpServerCertJob")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &McpServerCertJobReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})

		It("should set Ready condition to True and owner reference when catalog exists", func() {
			const resourceName = "test-job-with-catalog"
			const catalogName = "test-catalog"
			ctx := context.Background()
			namespacedName := types.NamespacedName{Name: resourceName, Namespace: "default"}

			// Create the referenced McpCatalog
			catalog := &mcpv1alpha1.McpCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name:      catalogName,
					Namespace: "default",
				},
				Spec: mcpv1alpha1.McpCatalogSpec{
					Description:   "desc",
					ImageRegistry: "quay.io/test",
				},
			}
			Expect(k8sClient.Create(ctx, catalog)).To(Succeed())
			DeferCleanup(func() {
				_ = k8sClient.Delete(ctx, catalog)
			})

			// Create the McpServerCertJob with the catalog label
			job := &mcpv1alpha1.McpServerCertJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
					Labels:    map[string]string{"mcp.opendatahub.io/mcpcatalog": catalogName},
				},
			}
			Expect(k8sClient.Create(ctx, job)).To(Succeed())
			DeferCleanup(func() {
				_ = k8sClient.Delete(ctx, job)
			})

			// Reconcile
			reconciler := &McpServerCertJobReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())

			// Check Ready condition and owner reference
			Eventually(func(g Gomega) {
				updated := &mcpv1alpha1.McpServerCertJob{}
				err := k8sClient.Get(ctx, namespacedName, updated)
				g.Expect(err).NotTo(HaveOccurred())
				cond := meta.FindStatusCondition(updated.Status.Conditions, "Ready")
				g.Expect(cond).NotTo(BeNil())
				g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))
				g.Expect(cond.Reason).To(Equal("ValidationSucceeded"))
				g.Expect(cond.Message).To(Equal("McpServer spec is valid"))
				g.Expect(updated.OwnerReferences).To(HaveLen(1))
				g.Expect(updated.OwnerReferences[0].Kind).To(Equal("McpCatalog"))
				g.Expect(updated.OwnerReferences[0].Name).To(Equal(catalogName))
			}).Should(Succeed())
		})

		It("should set Ready condition to False and no owner reference when catalog does not exist", func() {
			const resourceName = "test-job-without-catalog"
			const catalogName = "non-existent-catalog"
			ctx := context.Background()
			namespacedName := types.NamespacedName{Name: resourceName, Namespace: "default"}

			// Create the McpServerCertJob with a non-existent catalog label
			job := &mcpv1alpha1.McpServerCertJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
					Labels:    map[string]string{"mcp.opendatahub.io/mcpcatalog": catalogName},
				},
			}
			Expect(k8sClient.Create(ctx, job)).To(Succeed())
			DeferCleanup(func() {
				_ = k8sClient.Delete(ctx, job)
			})

			// Reconcile
			reconciler := &McpServerCertJobReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())

			// Check Ready condition and owner reference
			Eventually(func(g Gomega) {
				updated := &mcpv1alpha1.McpServerCertJob{}
				err := k8sClient.Get(ctx, namespacedName, updated)
				g.Expect(err).NotTo(HaveOccurred())
				cond := meta.FindStatusCondition(updated.Status.Conditions, "Ready")
				g.Expect(cond).NotTo(BeNil())
				g.Expect(cond.Status).To(Equal(metav1.ConditionFalse))
				g.Expect(cond.Reason).To(Equal("ValidationFailed"))
				g.Expect(cond.Message).To(Equal("referenced catalog does not exist"))
				g.Expect(updated.OwnerReferences).To(BeEmpty())
			}).Should(Succeed())
		})
	})
})
