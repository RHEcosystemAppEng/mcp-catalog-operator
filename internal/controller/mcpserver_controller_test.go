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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("McpServer Controller", func() {
	// Helper function to create McpCatalog instances
	createMcpCatalog := func(name, namespace, description, imageRegistry string) *mcpv1alpha1.McpCatalog {
		return &mcpv1alpha1.McpCatalog{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: mcpv1alpha1.McpCatalogSpec{
				Description:   description,
				ImageRegistry: imageRegistry,
			},
		}
	}

	// Helper function to create McpServer instances
	createMcpServer := func(name, namespace, catalogName string, catalogNamespace *string) *mcpv1alpha1.McpServer {
		server := &mcpv1alpha1.McpServer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: mcpv1alpha1.McpServerSpec{
				CatalogRef: mcpv1alpha1.CatalogRef{
					Name: catalogName,
				},
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

		// Set namespace if provided
		if catalogNamespace != nil {
			server.Spec.CatalogRef.Namespace = catalogNamespace
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
			readyCondition := meta.FindStatusCondition(updatedServer.Status.Conditions, mcpv1alpha1.ConditionTypeReady)
			g.Expect(readyCondition).NotTo(BeNil())
			g.Expect(readyCondition.Status).To(Equal(expectedStatus))
			g.Expect(readyCondition.Reason).To(Equal(expectedReason))
			g.Expect(readyCondition.Message).To(Equal(expectedMessage))
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
			server := createMcpServer(resourceName, "default", catalogName, nil)
			Expect(k8sClient.Create(ctx, server)).To(Succeed())
		})

		AfterEach(func() {
			cleanupResource("McpServer", typeNamespacedName, &mcpv1alpha1.McpServer{})
			cleanupResource("McpCatalog", catalogNamespacedName, &mcpv1alpha1.McpCatalog{})
		})

		It("should successfully reconcile and set Ready condition to True", func() {
			reconcileAndValidateStatus(typeNamespacedName, metav1.ConditionTrue, mcpv1alpha1.ConditionReasonValidationSucceeded, mcpv1alpha1.ValidationMessageServerSuccess)
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
			server := createMcpServer(resourceName, "default", catalogName, nil)
			Expect(k8sClient.Create(ctx, server)).To(Succeed())
		})

		AfterEach(func() {
			cleanupResource("McpServer", typeNamespacedName, &mcpv1alpha1.McpServer{})
		})

		It("should set Ready condition to False when catalog does not exist", func() {
			reconcileAndValidateStatus(typeNamespacedName, metav1.ConditionFalse, mcpv1alpha1.ConditionReasonValidationFailed, mcpv1alpha1.ValidationMessageCatalogNotFound)
		})
	})

	Context("When reconciling a resource with catalog in different namespace", func() {
		const resourceName = "test-server-cross-namespace"
		const catalogName = "test-catalog-cross-namespace"
		catalogNamespace := "catalog-namespace"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		catalogNamespacedName := types.NamespacedName{
			Name:      catalogName,
			Namespace: catalogNamespace,
		}

		BeforeEach(func() {
			By("creating the referenced McpCatalog in different namespace")
			// Create the namespace first
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: catalogNamespace,
				},
			}
			Expect(k8sClient.Create(ctx, namespace)).To(Succeed())

			catalog := createMcpCatalog(catalogName, catalogNamespace, "Test catalog in different namespace", "test-registry.example.com")
			Expect(k8sClient.Create(ctx, catalog)).To(Succeed())

			By("creating the McpServer resource with cross-namespace catalog reference")
			server := createMcpServer(resourceName, "default", catalogName, &catalogNamespace)
			Expect(k8sClient.Create(ctx, server)).To(Succeed())
		})

		AfterEach(func() {
			cleanupResource("McpServer", typeNamespacedName, &mcpv1alpha1.McpServer{})
			cleanupResource("McpCatalog", catalogNamespacedName, &mcpv1alpha1.McpCatalog{})

			By("Cleanup the namespace")
			namespace := &corev1.Namespace{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: catalogNamespace}, namespace)
			if err == nil {
				Expect(k8sClient.Delete(ctx, namespace)).To(Succeed())
			}
		})

		It("should successfully reconcile with catalog in different namespace", func() {
			reconcileAndValidateStatus(typeNamespacedName, metav1.ConditionTrue, mcpv1alpha1.ConditionReasonValidationSucceeded, mcpv1alpha1.ValidationMessageServerSuccess)
		})
	})
})
