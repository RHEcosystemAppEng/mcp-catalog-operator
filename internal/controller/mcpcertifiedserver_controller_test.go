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
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
)

// Test helper functions
func createMcpCertifiedServerSpec(options ...func(*mcpv1alpha1.McpCertifiedServerSpec)) mcpv1alpha1.McpCertifiedServerSpec {
	spec := mcpv1alpha1.McpCertifiedServerSpec{
		McpServer: mcpv1alpha1.McpCertifiedServerServerSpec{
			Type: "",
		},
	}

	// Apply options to customize the spec
	for _, option := range options {
		option(&spec)
	}

	return spec
}

func withContainerImage() func(*mcpv1alpha1.McpCertifiedServerSpec) {
	return func(spec *mcpv1alpha1.McpCertifiedServerSpec) {
		if spec.McpServer.Type != mcpv1alpha1.ServerTypeContainer {
			spec.McpServer.Type = mcpv1alpha1.ServerTypeContainer
		}
		if spec.McpServer.Container == nil {
			spec.McpServer.Container = &mcpv1alpha1.ContainerServerSpec{
				Command: "mcp-registry",
				Args:    []string{"--port", "8080"},
				EnvVars: []string{"TEST=value"},
			}
		}
		spec.McpServer.Container.Image = "custom:latest"
	}
}
func withRemoteUri() func(*mcpv1alpha1.McpCertifiedServerSpec) {
	return func(spec *mcpv1alpha1.McpCertifiedServerSpec) {
		if spec.McpServer.Type != mcpv1alpha1.ServerTypeRemote {
			spec.McpServer.Type = mcpv1alpha1.ServerTypeRemote
		}
		if spec.McpServer.Remote == nil {
			spec.McpServer.Remote = &mcpv1alpha1.RemoteServerSpec{
				TransportType: "http",
			}
		}
		spec.McpServer.Remote.URL = "http://localhost:8080"
	}
}

func createMcpCertifiedServer(name string, spec mcpv1alpha1.McpCertifiedServerSpec, catalogName string) *mcpv1alpha1.McpCertifiedServer {
	labels := make(map[string]string)
	if catalogName != "" {
		labels[McpCatalogNameLabel] = catalogName
	}

	return &mcpv1alpha1.McpCertifiedServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Labels:    labels,
		},
		Spec: spec,
	}
}

func createMcpCertifiedServerWithoutLabels(name, namespace string, spec mcpv1alpha1.McpCertifiedServerSpec) *mcpv1alpha1.McpCertifiedServer {
	return &mcpv1alpha1.McpCertifiedServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    nil,
		},
		Spec: spec,
	}
}

func createAndExpectResource(ctx context.Context, resource *mcpv1alpha1.McpCertifiedServer, shouldSucceed bool) {
	err := k8sClient.Create(ctx, resource)
	if shouldSucceed {
		Expect(err).To(Succeed())
	} else {
		Expect(err).To(HaveOccurred())
	}
}

func cleanupResource(ctx context.Context, namespacedName types.NamespacedName) {
	resource := &mcpv1alpha1.McpCertifiedServer{}
	err := k8sClient.Get(ctx, namespacedName, resource)
	if err == nil {
		Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
	}
}

func createAndExpectMcpCatalog(ctx context.Context, catalog *mcpv1alpha1.McpCatalog, shouldSucceed bool) {
	err := k8sClient.Create(ctx, catalog)
	if shouldSucceed {
		Expect(err).To(Succeed())
	} else {
		Expect(err).To(HaveOccurred())
	}
}

func cleanupMcpCatalog(ctx context.Context, namespacedName types.NamespacedName) {
	catalog := &mcpv1alpha1.McpCatalog{}
	err := k8sClient.Get(ctx, namespacedName, catalog)
	if err == nil {
		Expect(k8sClient.Delete(ctx, catalog)).To(Succeed())
	}
}

func hasOwnerReference(resource *mcpv1alpha1.McpCertifiedServer, ownerName, ownerKind string) bool {
	for _, ref := range resource.GetOwnerReferences() {
		if ref.Name == ownerName && ref.Kind == ownerKind {
			return true
		}
	}
	return false
}

var _ = Describe("McpCertifiedServer Controller", func() {
	const namespace = "default"
	ctx := context.Background()

	Context("When reconciling a resource", func() {
		DescribeTable("different McpCertifiedServer configurations",
			func(resourceName string, specOptions []func(*mcpv1alpha1.McpCertifiedServerSpec), shouldCreateSucceed bool, shouldReconcileSucceed bool) {
				typeNamespacedName := types.NamespacedName{
					Name:      resourceName,
					Namespace: namespace,
				}

				spec := createMcpCertifiedServerSpec(specOptions...)
				resource := createMcpCertifiedServer(resourceName, spec, "")

				By("creating the custom resource for the Kind McpCertifiedServer")
				createAndExpectResource(ctx, resource, shouldCreateSucceed)

				if shouldCreateSucceed {
					defer cleanupResource(ctx, typeNamespacedName)

					By("Reconciling the created resource")
					controllerReconciler := &McpCertifiedServerReconciler{
						Client: k8sClient,
						Scheme: k8sClient.Scheme(),
					}

					_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: typeNamespacedName,
					})

					if shouldReconcileSucceed {
						Expect(err).NotTo(HaveOccurred())
					} else {
						Expect(err).To(HaveOccurred())
					}
				}
			},
			Entry("with custom image",
				"test-custom-image",
				[]func(*mcpv1alpha1.McpCertifiedServerSpec){withContainerImage()},
				true,  // creation should succeed
				false, // reconcile should fail (no registry)
			),
			Entry("with remote URI",
				"test-remote-uri",
				[]func(*mcpv1alpha1.McpCertifiedServerSpec){withRemoteUri()},
				true,  // creation should succeed
				false, // reconcile should fail (no catalog)
			),
		)
	})

	Context("When testing specific scenarios", func() {
		It("should fail to reconcile when catalog is missing", func() {
			resourceName := "test-no-catalog"
			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: namespace,
			}

			By("creating the custom resource for the Kind McpCertifiedServer with no McpRegistry")
			spec := createMcpCertifiedServerSpec(withContainerImage())
			resource := createMcpCertifiedServer(resourceName, spec, "")

			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			defer cleanupResource(ctx, typeNamespacedName)

			By("Reconciling the created resource")
			controllerReconciler := &McpCertifiedServerReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When testing owner reference behavior", func() {
		It("should set owner reference when annotated McpCatalog exists", func() {
			catalogName := "test-catalog"
			catalogNamespace := namespace
			resourceName := "test-with-catalog"

			catalogNamespacedName := types.NamespacedName{
				Name:      catalogName,
				Namespace: catalogNamespace,
			}
			resourceNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: namespace,
			}

			By("creating the McpCatalog first")
			catalog := createMcpCatalog(catalogName, catalogNamespace, "Test catalog", "quay.io/test")
			createAndExpectMcpCatalog(ctx, catalog, true)
			defer cleanupMcpCatalog(ctx, catalogNamespacedName)

			By("creating the McpCertifiedServer with catalog labels")
			spec := createMcpCertifiedServerSpec(withContainerImage())
			resource := createMcpCertifiedServer(resourceName, spec, catalogName)

			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			defer cleanupResource(ctx, resourceNamespacedName)

			By("Reconciling the created resource")
			controllerReconciler := &McpCertifiedServerReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: resourceNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying that owner reference is set")
			updatedResource := &mcpv1alpha1.McpCertifiedServer{}
			Expect(k8sClient.Get(ctx, resourceNamespacedName, updatedResource)).To(Succeed())
			Expect(hasOwnerReference(updatedResource, catalogName, "McpCatalog")).To(BeTrue())
		})

		It("should not set owner reference and fail reconciliation when annotated McpCatalog does not exist", func() {
			resourceName := "test-with-missing-catalog"
			resourceNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: namespace,
			}

			By("creating the McpCertifiedServer with non-existent catalog labels")
			spec := createMcpCertifiedServerSpec(withContainerImage())
			resource := createMcpCertifiedServer(resourceName, spec, "non-existent-catalog")

			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			defer cleanupResource(ctx, resourceNamespacedName)

			By("Reconciling the created resource")
			controllerReconciler := &McpCertifiedServerReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: resourceNamespacedName,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get McpCatalog"))

			By("Verifying that owner reference is not set")
			updatedResource := &mcpv1alpha1.McpCertifiedServer{}
			Expect(k8sClient.Get(ctx, resourceNamespacedName, updatedResource)).To(Succeed())
			Expect(hasOwnerReference(updatedResource, "non-existent-catalog", "McpCatalog")).To(BeFalse())
		})

		It("should not set owner reference and fail reconciliation when catalog label is missing", func() {
			resourceName := "test-no-catalog-label"
			resourceNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: namespace,
			}

			By("creating the McpCertifiedServer without catalog labels")
			spec := createMcpCertifiedServerSpec(withContainerImage())
			resource := createMcpCertifiedServerWithoutLabels(resourceName, namespace, spec)

			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			defer cleanupResource(ctx, resourceNamespacedName)

			By("Reconciling the created resource")
			controllerReconciler := &McpCertifiedServerReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: resourceNamespacedName,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no labels found on object"))

			By("Verifying that no owner references are set")
			updatedResource := &mcpv1alpha1.McpCertifiedServer{}
			Expect(k8sClient.Get(ctx, resourceNamespacedName, updatedResource)).To(Succeed())
			Expect(updatedResource.GetOwnerReferences()).To(BeEmpty())
		})
	})
})
