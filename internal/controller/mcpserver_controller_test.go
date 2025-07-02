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
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("McpServer Controller", func() {
	// Helper function to create McpServer instances
	createMcpServer := func(name, namespace, catalogName string) *mcpv1alpha1.McpServer {
		labels := make(map[string]string)
		if catalogName != "" {
			labels[McpCatalogLabel] = catalogName
		}

		server := &mcpv1alpha1.McpServer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    labels,
			},
			Spec: mcpv1alpha1.McpServerSpec{
				ServerDetail: mcpv1alpha1.ServerDetail{
					Server: mcpv1alpha1.Server{
						ID:          "test-server",
						Name:        "Test Server",
						Description: "A test server",
						Repository: mcpv1alpha1.Repository{
							URL:    "https://github.com/test/server",
							Source: "github",
							ID:     "test-repo",
						},
						VersionDetail: mcpv1alpha1.VersionDetail{
							Version:     "1.0.0",
							ReleaseDate: "2024-01-01",
							IsLatest:    true,
						},
					},
				},
			},
		}

		return server
	}

	// Helper function to create controller reconciler
	createControllerReconciler := func() *McpServerReconciler {
		return &McpServerReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}
	}

	// Helper function to reconcile and validate status
	reconcileAndValidateStatus := func(namespacedName types.NamespacedName, expectedStatus metav1.ConditionStatus, expectedReason, expectedMessage string) {
		By("Reconciling the created resource")
		controllerReconciler := createControllerReconciler()

		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: namespacedName,
		})
		Expect(err).NotTo(HaveOccurred())

		// Use Eventually to wait for the status to be updated
		Eventually(func(g Gomega) {
			updatedServer := &mcpv1alpha1.McpServer{}
			err := k8sClient.Get(ctx, namespacedName, updatedServer)
			g.Expect(err).NotTo(HaveOccurred())
			readyCondition := meta.FindStatusCondition(updatedServer.Status.Conditions, ConditionTypeReady)
			g.Expect(readyCondition).NotTo(BeNil())
			g.Expect(readyCondition.Status).To(Equal(expectedStatus))
			g.Expect(readyCondition.Reason).To(Equal(expectedReason))
			g.Expect(readyCondition.Message).To(Equal(expectedMessage))

			if expectedStatus == metav1.ConditionTrue {
				catalogName := updatedServer.GetLabels()[McpCatalogLabel]
				g.Expect(updatedServer.OwnerReferences).To(HaveLen(1))
				g.Expect(updatedServer.OwnerReferences[0].Kind).To(Equal("McpCatalog"))
				g.Expect(updatedServer.OwnerReferences[0].Name).To(Equal(catalogName))
			} else {
				g.Expect(updatedServer.OwnerReferences).To(BeEmpty())
			}
		}).Should(Succeed())
	}

	// Helper function to cleanup resources
	cleanupResource := func(resourceType string, namespacedName types.NamespacedName, resource client.Object) {
		By("Cleanup the " + resourceType + " resource")
		err := k8sClient.Get(ctx, namespacedName, resource)
		if err == nil {
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		}
	}

	Context("When reconciling a resource with existing catalog", func() {
		const resourceName = "test-server-with-catalog"
		const catalogName = "test-catalog"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		catalogNamespacedName := types.NamespacedName{
			Name:      catalogName,
			Namespace: "default",
		}

		BeforeEach(func() {
			By("creating the referenced McpCatalog")
			catalog := createMcpCatalog(catalogName, "default", "Test catalog for server validation", "test-registry.example.com")
			Expect(k8sClient.Create(ctx, catalog)).To(Succeed())

			By("creating the McpServer resource")
			server := createMcpServer(resourceName, "default", catalogName)
			Expect(k8sClient.Create(ctx, server)).To(Succeed())
		})

		AfterEach(func() {
			cleanupResource("McpServer", typeNamespacedName, &mcpv1alpha1.McpServer{})
			cleanupResource("McpCatalog", catalogNamespacedName, &mcpv1alpha1.McpCatalog{})
		})

		It("should successfully reconcile and set Ready condition to True", func() {
			reconcileAndValidateStatus(typeNamespacedName, metav1.ConditionTrue, ConditionReasonValidationSucceeded, ValidationMessageServerSuccess)
		})
	})

	Context("When reconciling a resource without existing catalog", func() {
		const resourceName = "test-server-without-catalog"
		const catalogName = "non-existent-catalog"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}

		BeforeEach(func() {
			By("creating the McpServer resource with non-existent catalog reference")
			server := createMcpServer(resourceName, "default", catalogName)
			Expect(k8sClient.Create(ctx, server)).To(Succeed())
		})

		AfterEach(func() {
			cleanupResource("McpServer", typeNamespacedName, &mcpv1alpha1.McpServer{})
		})

		It("should set Ready condition to False when catalog does not exist", func() {
			reconcileAndValidateStatus(typeNamespacedName, metav1.ConditionFalse, ConditionReasonValidationFailed, ValidationMessageCatalogNotFound)
		})
	})
})
