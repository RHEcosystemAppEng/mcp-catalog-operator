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

	mcpv1 "github.com/dmartinol/mcp-registry-operator/api/v1"
)

// Test helper functions
func createMcpCertifiedServerSpec(options ...func(*mcpv1.McpCertifiedServerSpec)) mcpv1.McpCertifiedServerSpec {
	namespace := "default"
	spec := mcpv1.McpCertifiedServerSpec{
		CatalogRef: mcpv1.CatalogRef{
			Name:      "test-registry",
			Namespace: &namespace,
		},
		Description:  "Test server",
		Provider:     "test",
		License:      "MIT",
		Competencies: []string{"test"},
		McpServer: mcpv1.McpCertifiedServerServerSpec{
			Image:   "quay.io/ecosystem-appeng/mcp-registry:amd64-0.1",
			Command: "mcp-registry",
			Args:    []string{"--port", "8080"},
			EnvVars: []string{"TEST=value"},
		},
	}

	// Apply options to customize the spec
	for _, option := range options {
		option(&spec)
	}

	return spec
}

func withCatalogRef(name, namespace string) func(*mcpv1.McpCertifiedServerSpec) {
	return func(spec *mcpv1.McpCertifiedServerSpec) {
		spec.CatalogRef.Name = name
		if namespace != "" {
			spec.CatalogRef.Namespace = &namespace
		}
	}
}

func withCompetencies(competencies []string) func(*mcpv1.McpCertifiedServerSpec) {
	return func(spec *mcpv1.McpCertifiedServerSpec) {
		spec.Competencies = competencies
	}
}

func withImage(image string) func(*mcpv1.McpCertifiedServerSpec) {
	return func(spec *mcpv1.McpCertifiedServerSpec) {
		spec.McpServer.Image = image
	}
}

func createMcpCertifiedServer(name, namespace string, spec mcpv1.McpCertifiedServerSpec) *mcpv1.McpCertifiedServer {
	return &mcpv1.McpCertifiedServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: spec,
	}
}

func createAndExpectResource(ctx context.Context, resource *mcpv1.McpCertifiedServer, shouldSucceed bool) {
	err := k8sClient.Create(ctx, resource)
	if shouldSucceed {
		Expect(err).To(Succeed())
	} else {
		Expect(err).To(HaveOccurred())
	}
}

func cleanupResource(ctx context.Context, namespacedName types.NamespacedName) {
	resource := &mcpv1.McpCertifiedServer{}
	err := k8sClient.Get(ctx, namespacedName, resource)
	if err == nil {
		Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
	}
}

var _ = Describe("McpCertifiedServer Controller", func() {
	const namespace = "default"
	ctx := context.Background()

	Context("When reconciling a resource", func() {
		DescribeTable("different McpCertifiedServer configurations",
			func(resourceName string, specOptions []func(*mcpv1.McpCertifiedServerSpec), shouldCreateSucceed bool, shouldReconcileSucceed bool) {
				typeNamespacedName := types.NamespacedName{
					Name:      resourceName,
					Namespace: namespace,
				}

				spec := createMcpCertifiedServerSpec(specOptions...)
				resource := createMcpCertifiedServer(resourceName, namespace, spec)

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
			Entry("with missing registry reference",
				"test-missing-registry",
				[]func(*mcpv1.McpCertifiedServerSpec){withCatalogRef("non-existent-registry", namespace)},
				true,  // creation should succeed
				false, // reconcile should fail
			),
			Entry("with different competencies",
				"test-different-competencies",
				[]func(*mcpv1.McpCertifiedServerSpec){withCompetencies([]string{"ai", "search", "tools"})},
				true,  // creation should succeed
				false, // reconcile should fail (no registry)
			),
			Entry("with custom image",
				"test-custom-image",
				[]func(*mcpv1.McpCertifiedServerSpec){withImage("custom:latest")},
				true,  // creation should succeed
				false, // reconcile should fail (no registry)
			),
		)
	})

	Context("When testing specific scenarios", func() {
		It("should fail to reconcile when registry is missing", func() {
			resourceName := "test-no-registry"
			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: namespace,
			}

			By("creating the custom resource for the Kind McpCertifiedServer with no McpRegistry")
			spec := createMcpCertifiedServerSpec(withCatalogRef("missing-registry", namespace))
			resource := createMcpCertifiedServer(resourceName, namespace, spec)

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
})
