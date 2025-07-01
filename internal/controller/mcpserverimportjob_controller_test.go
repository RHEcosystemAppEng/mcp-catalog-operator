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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
)

// Helper function to create McpServerImportJob instances
var createMcpServerImportJob = func(name, namespace, registryURI, catalogName string, nameFilter *string, maxServers *int) *mcpv1alpha1.McpServerImportJob {
	labels := make(map[string]string)
	if catalogName != "" {
		labels[McpCatalogNameLabel] = catalogName
	}

	return &mcpv1alpha1.McpServerImportJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: mcpv1alpha1.McpServerImportJobSpec{
			RegistryURI: registryURI,
			NameFilter:  nameFilter,
			MaxServers:  maxServers,
		},
	}
}

var _ = Describe("McpServerImportJob Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		mcpserverimportjob := &mcpv1alpha1.McpServerImportJob{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind McpServerImportJob")
			err := k8sClient.Get(ctx, typeNamespacedName, mcpserverimportjob)
			if err != nil && errors.IsNotFound(err) {
				catalog := createMcpCatalog("test-catalog", "default", "Test catalog", "quay.io/test")
				Expect(k8sClient.Create(ctx, catalog)).To(Succeed())

				resource := createMcpServerImportJob(
					resourceName,
					"default",
					"http://mcp-registry.com",
					"test-catalog",
					nil, // nameFilter
					nil, // maxServers
				)
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &mcpv1alpha1.McpServerImportJob{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance McpServerImportJob")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			// Verify that the Job is deleted (due to owner reference) or manually delete it
			Eventually(func(g Gomega) {
				jobList := &batchv1.JobList{}
				err := k8sClient.List(ctx, jobList, client.InNamespace("default"), client.MatchingLabels(map[string]string{
					McpServerImportJobLabel: resourceName,
				}))
				fmt.Fprintf(GinkgoWriter, "Debug: Found %d jobs, error: %v", len(jobList.Items), err)

				if len(jobList.Items) > 0 {
					job := jobList.Items[0]
					fmt.Fprintf(GinkgoWriter, "Debug: Job owner references: %v", job.OwnerReferences)
					fmt.Fprintf(GinkgoWriter, "Debug: Job name: %s", job.Name)

					// If Job still exists, manually delete it
					By("Manually deleting Job since owner reference deletion didn't work")
					err = k8sClient.Delete(ctx, &job, client.PropagationPolicy(metav1.DeletePropagationBackground))
					g.Expect(err).NotTo(HaveOccurred())
				}

				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(jobList.Items).To(BeEmpty())
			}, 10*time.Second, 100*time.Millisecond).Should(Succeed())

			// Verify that SA, Role, and RoleBinding remain (they should not be deleted)
			Eventually(func(g Gomega) {
				// Check ServiceAccount still exists
				sa := &corev1.ServiceAccount{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      McpServerImporterServiceAccountName,
					Namespace: "default",
				}, sa)
				g.Expect(err).NotTo(HaveOccurred())

				// Check Role still exists
				role := &rbacv1.Role{}
				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      McpServerImporterRoleName,
					Namespace: "default",
				}, role)
				g.Expect(err).NotTo(HaveOccurred())

				// Check RoleBinding still exists
				roleBinding := &rbacv1.RoleBinding{}
				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      McpServerImporterRoleBindingName,
					Namespace: "default",
				}, roleBinding)
				g.Expect(err).NotTo(HaveOccurred())
			}, 5*time.Second, 100*time.Millisecond).Should(Succeed())
		})

		It("should successfully reconcile the resource and create expected resources", func() {
			By("Reconciling the created resource")
			controllerReconciler := &McpServerImportJobReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Wait for the Job to be created
			Eventually(func(g Gomega) {
				jobList := &batchv1.JobList{}
				err := k8sClient.List(ctx, jobList, client.InNamespace("default"), client.MatchingLabels(map[string]string{
					McpServerImportJobLabel: resourceName,
				}))
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(jobList.Items).To(HaveLen(1))

				job := jobList.Items[0]
				fmt.Fprintf(GinkgoWriter, "Debug: Job created with owner references: %v", job.OwnerReferences)
				fmt.Fprintf(GinkgoWriter, "Debug: Job name: %s", job.Name)
				// Verify Job has expected labels
				g.Expect(job.Labels).To(HaveKeyWithValue(McpServerImportJobLabel, resourceName))
				g.Expect(job.Labels).To(HaveKeyWithValue(McpCatalogNameLabel, "test-catalog"))

				// Verify Job spec
				g.Expect(job.Spec.Template.Spec.ServiceAccountName).To(Equal(McpServerImporterServiceAccountName))
				g.Expect(job.Spec.Template.Spec.Containers).To(HaveLen(1))
				g.Expect(job.Spec.Template.Spec.Containers[0].Name).To(Equal(McpServerImporterContainerName))
				g.Expect(job.Spec.Template.Spec.Containers[0].Image).To(Equal(McpServerImporterImage))

				// Verify environment variables
				envVars := job.Spec.Template.Spec.Containers[0].Env
				g.Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "CATALOG_NAME", Value: "test-catalog"}))
				g.Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "REGISTRY_URL", Value: "http://mcp-registry.com"}))
				g.Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "IMPORT_JOB_NAME", Value: resourceName}))
				g.Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "MAX_SERVERS", Value: "10"}))
			}, 10*time.Second, 100*time.Millisecond).Should(Succeed())

			// Verify ServiceAccount was created
			Eventually(func(g Gomega) {
				sa := &corev1.ServiceAccount{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      McpServerImporterServiceAccountName,
					Namespace: "default",
				}, sa)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(sa.Name).To(Equal(McpServerImporterServiceAccountName))
			}, 5*time.Second, 100*time.Millisecond).Should(Succeed())

			// Verify Role was created
			Eventually(func(g Gomega) {
				role := &rbacv1.Role{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      McpServerImporterRoleName,
					Namespace: "default",
				}, role)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(role.Name).To(Equal(McpServerImporterRoleName))
				g.Expect(role.Rules).To(HaveLen(2))
				g.Expect(role.Rules[0].APIGroups).To(ContainElement("mcp.opendatahub.io"))
				g.Expect(role.Rules[0].Resources).To(ContainElement("mcpservers"))
				g.Expect(role.Rules[1].APIGroups).To(ContainElement(""))
				g.Expect(role.Rules[1].Resources).To(ContainElement("configmaps"))
			}, 5*time.Second, 100*time.Millisecond).Should(Succeed())

			// Verify RoleBinding was created
			Eventually(func(g Gomega) {
				roleBinding := &rbacv1.RoleBinding{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      McpServerImporterRoleBindingName,
					Namespace: "default",
				}, roleBinding)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(roleBinding.Name).To(Equal(McpServerImporterRoleBindingName))
				g.Expect(roleBinding.Subjects).To(HaveLen(1))
				g.Expect(roleBinding.Subjects[0].Name).To(Equal(McpServerImporterServiceAccountName))
				g.Expect(roleBinding.RoleRef.Name).To(Equal(McpServerImporterRoleName))
			}, 5*time.Second, 100*time.Millisecond).Should(Succeed())
		})
	})
})
