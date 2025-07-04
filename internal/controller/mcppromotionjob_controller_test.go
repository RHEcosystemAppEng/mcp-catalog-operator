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
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
)

var _ = Describe("McpPromotionJob Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"
		const stagingAreaName = "test-stagingarea"
		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}

		BeforeEach(func() {
			By("cleaning up any existing resources")
			_ = k8sClient.Delete(ctx, &mcpv1alpha1.McpPromotionJob{ObjectMeta: metav1.ObjectMeta{Name: resourceName, Namespace: "default"}})
			_ = k8sClient.Delete(ctx, &mcpv1alpha1.McpStagingArea{ObjectMeta: metav1.ObjectMeta{Name: stagingAreaName, Namespace: "default"}})
		})

		AfterEach(func() {
			By("Cleanup the specific resource instance McpPromotionJob")
			_ = k8sClient.Delete(ctx, &mcpv1alpha1.McpPromotionJob{ObjectMeta: metav1.ObjectMeta{Name: resourceName, Namespace: "default"}})
			_ = k8sClient.Delete(ctx, &mcpv1alpha1.McpStagingArea{ObjectMeta: metav1.ObjectMeta{Name: stagingAreaName, Namespace: "default"}})
		})

		It("should successfully reconcile the resource and initialize status when McpStagingArea exists", func() {
			By("creating the referenced McpStagingArea")
			stagingArea := &mcpv1alpha1.McpStagingArea{
				ObjectMeta: metav1.ObjectMeta{
					Name:      stagingAreaName,
					Namespace: "default",
				},
			}
			Expect(k8sClient.Create(ctx, stagingArea)).To(Succeed())

			By("creating the McpPromotionJob with the correct label and servers")
			promotionJob := &mcpv1alpha1.McpPromotionJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
					Labels: map[string]string{
						McpStagingAreaLabel: stagingAreaName,
					},
				},
				Spec: mcpv1alpha1.McpPromotionJobSpec{
					Servers: []string{"server1", "server2"},
				},
			}
			Expect(k8sClient.Create(ctx, promotionJob)).To(Succeed())

			controllerReconciler := &McpPromotionJobReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("validating the status.ServerPromotions field")
			updated := &mcpv1alpha1.McpPromotionJob{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, updated)).To(Succeed())
			Expect(updated.Status.ServerPromotions).To(HaveLen(2))
			for i, name := range []string{"server1", "server2"} {
				sp := updated.Status.ServerPromotions[i]
				Expect(sp.Name).To(Equal(name))
				Expect(sp.PromotionStatus).To(Equal(mcpv1alpha1.PromotionStatusPlanned))
				Expect(sp.OriginalImage).To(BeEmpty())
				Expect(sp.DestinationImage).To(BeEmpty())
			}
		})

		It("should fail to reconcile if the referenced McpStagingArea does not exist", func() {
			By("creating the McpPromotionJob with a non-existent staging area label")
			promotionJob := &mcpv1alpha1.McpPromotionJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
					Labels: map[string]string{
						McpStagingAreaLabel: "nonexistent-stagingarea",
					},
				},
				Spec: mcpv1alpha1.McpPromotionJobSpec{
					Servers: []string{"server1"},
				},
			}
			Expect(k8sClient.Create(ctx, promotionJob)).To(Succeed())

			controllerReconciler := &McpPromotionJobReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())
			Expect(fmt.Sprintf("%v", err)).To(ContainSubstring("failed to get"))
		})
	})
})
