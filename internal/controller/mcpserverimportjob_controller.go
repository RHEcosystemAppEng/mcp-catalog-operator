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
	"strconv"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// McpServerImportJobReconciler reconciles a McpServerImportJob object
type McpServerImportJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// initializeJob initializes missing fields in the McpServerImportJob
// Returns an error if initialization fails, and a boolean indicating if the job needs to be started
func (r *McpServerImportJobReconciler) initializeJob(ctx context.Context, mcpServerImportJob *mcpv1alpha1.McpServerImportJob) (bool, error) {
	log := logf.FromContext(ctx)

	// Initialize missing fields in spec
	needsUpdate := false
	needsStart := false
	defaultNameFilter := ""
	defaultMaxServers := DefaultMaxServers

	// Initialize NameFilter to empty string if not set
	if mcpServerImportJob.Spec.NameFilter == nil {
		mcpServerImportJob.Spec.NameFilter = &defaultNameFilter
		needsUpdate = true
	}

	// Initialize MaxServers to 10 if not set
	if mcpServerImportJob.Spec.MaxServers == nil {
		mcpServerImportJob.Spec.MaxServers = &defaultMaxServers
		needsUpdate = true
	}

	// Initialize status if not set
	if mcpServerImportJob.Status.Status == "" {
		mcpServerImportJob.Status.Status = mcpv1alpha1.ImportJobRunning
		needsUpdate = true
		needsStart = true
	}

	// Initialize ConfigMapName to empty string if not set
	if mcpServerImportJob.Status.ConfigMapName == "" {
		mcpServerImportJob.Status.ConfigMapName = ""
		needsUpdate = true
	}

	// Update the resource if any fields were initialized
	if needsUpdate {
		log.Info("Initializing McpServerImportJob fields",
			"name", mcpServerImportJob.Name,
			"namespace", mcpServerImportJob.Namespace)

		err := r.Status().Update(ctx, mcpServerImportJob)
		if err != nil {
			log.Error(err, "Failed to update McpServerImportJob status")
			return false, err
		}

		// Also update the spec if it was modified
		if mcpServerImportJob.Spec.NameFilter == nil || mcpServerImportJob.Spec.MaxServers == nil {
			err = r.Update(ctx, mcpServerImportJob)
			if err != nil {
				log.Error(err, "Failed to update McpServerImportJob spec")
				return false, err
			}
		}
	}

	return needsStart, nil
}

// createServiceAccount creates the mcpserver-importer ServiceAccount if it doesn't exist
func (r *McpServerImportJobReconciler) createServiceAccount(ctx context.Context, namespace string) error {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      McpServerImporterServiceAccountName,
			Namespace: namespace,
		},
	}

	err := r.Create(ctx, serviceAccount)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create ServiceAccount: %w", err)
	}

	return nil
}

// createRole creates the mcpserver-importer Role if it doesn't exist
func (r *McpServerImportJobReconciler) createRole(ctx context.Context, namespace string) error {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      McpServerImporterRoleName,
			Namespace: namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"mcp.opendatahub.io"},
				Resources: []string{"mcpservers"},
				Verbs:     []string{"get", "list", "create", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"create"},
			},
		},
	}

	err := r.Create(ctx, role)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create Role: %w", err)
	}

	return nil
}

// createRoleBinding creates the RoleBinding for the mcpserver-importer ServiceAccount if it doesn't exist
func (r *McpServerImportJobReconciler) createRoleBinding(ctx context.Context, namespace string) error {
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      McpServerImporterRoleBindingName,
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      McpServerImporterServiceAccountName,
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     McpServerImporterRoleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	err := r.Create(ctx, roleBinding)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create RoleBinding: %w", err)
	}

	return nil
}

// createImportJob creates the Kubernetes Job for importing MCP servers
func (r *McpServerImportJobReconciler) createImportJob(ctx context.Context, mcpServerImportJob *mcpv1alpha1.McpServerImportJob) error {
	log := logf.FromContext(ctx)

	// Get the catalog name from labels
	catalogName := mcpServerImportJob.Labels[McpCatalogNameLabel]
	if catalogName == "" {
		return fmt.Errorf("McpCatalogNameLabel not found on McpServerImportJob")
	}

	// Convert MaxServers to string
	maxServersStr := strconv.Itoa(DefaultMaxServers) // default value
	if mcpServerImportJob.Spec.MaxServers != nil {
		maxServersStr = strconv.Itoa(*mcpServerImportJob.Spec.MaxServers)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: McpServerImporterJobGenerateName,
			Namespace:    mcpServerImportJob.Namespace,
			Labels: map[string]string{
				McpServerImportJobLabel: mcpServerImportJob.Name,
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: McpServerImporterServiceAccountName,
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name:  McpServerImporterContainerName,
							Image: McpServerImporterImage,
							Env: []corev1.EnvVar{
								{
									Name:  "CATALOG_NAME",
									Value: catalogName,
								},
								{
									Name:  "REGISTRY_URL",
									Value: mcpServerImportJob.Spec.RegistryURI,
								},
								{
									Name:  "IMPORT_JOB_NAME",
									Value: mcpServerImportJob.Name,
								},
								{
									Name:  "MAX_SERVERS",
									Value: maxServersStr,
								},
							},
						},
					},
				},
			},
		},
	}

	// Set the owner reference to make the Job owned by the McpServerImportJob
	if err := ctrl.SetControllerReference(mcpServerImportJob, job, r.Scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	err := r.Create(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to create Job: %w", err)
	}

	log.Info("Created import Job", "jobName", job.Name, "namespace", job.Namespace)
	return nil
}

// startImportJob creates all necessary resources and starts the import job
func (r *McpServerImportJobReconciler) startImportJob(ctx context.Context, mcpServerImportJob *mcpv1alpha1.McpServerImportJob) error {
	log := logf.FromContext(ctx)

	// Create ServiceAccount
	if err := r.createServiceAccount(ctx, mcpServerImportJob.Namespace); err != nil {
		log.Error(err, "Failed to create ServiceAccount")
		return err
	}

	// Create Role
	if err := r.createRole(ctx, mcpServerImportJob.Namespace); err != nil {
		log.Error(err, "Failed to create Role")
		return err
	}

	// Create RoleBinding
	if err := r.createRoleBinding(ctx, mcpServerImportJob.Namespace); err != nil {
		log.Error(err, "Failed to create RoleBinding")
		return err
	}

	// Create the import Job
	if err := r.createImportJob(ctx, mcpServerImportJob); err != nil {
		log.Error(err, "Failed to create import Job")
		return err
	}

	log.Info("Successfully started import job", "name", mcpServerImportJob.Name, "namespace", mcpServerImportJob.Namespace)
	return nil
}

// checkJobStatus checks the status of the Job owned by this McpServerImportJob and updates the status accordingly
func (r *McpServerImportJobReconciler) checkJobStatus(ctx context.Context, mcpServerImportJob *mcpv1alpha1.McpServerImportJob) error {
	log := logf.FromContext(ctx)

	// List Jobs owned by this McpServerImportJob
	jobList := &batchv1.JobList{}
	err := r.List(ctx, jobList, client.InNamespace(mcpServerImportJob.Namespace), client.MatchingLabels(map[string]string{
		McpServerImportJobLabel: mcpServerImportJob.Name,
	}))
	if err != nil {
		return fmt.Errorf("failed to list Jobs: %w", err)
	}

	if len(jobList.Items) == 0 {
		log.Info("No Jobs found for McpServerImportJob", "name", mcpServerImportJob.Name, "namespace", mcpServerImportJob.Namespace)
		return nil
	}

	// Get the first Job (there should only be one)
	job := &jobList.Items[0]
	log.Info("Found Job for McpServerImportJob", "jobName", job.Name, "jobStatus", job.Status.Conditions)

	// Check Job status
	needsUpdate := false
	var newStatus mcpv1alpha1.ImportJobStatus

	if len(job.Status.Conditions) > 0 {
		latestCondition := job.Status.Conditions[0]
		for _, condition := range job.Status.Conditions {
			if condition.LastTransitionTime.After(latestCondition.LastTransitionTime.Time) {
				latestCondition = condition
			}
		}

		switch latestCondition.Type {
		case batchv1.JobComplete:
			if latestCondition.Status == corev1.ConditionTrue {
				newStatus = mcpv1alpha1.ImportJobCompleted
				needsUpdate = true
				log.Info("Job completed successfully", "jobName", job.Name)
			}
		case batchv1.JobFailed:
			if latestCondition.Status == corev1.ConditionTrue {
				newStatus = mcpv1alpha1.ImportJobFailed
				needsUpdate = true
				log.Info("Job failed", "jobName", job.Name, "reason", latestCondition.Reason, "message", latestCondition.Message)
			}
		}
	}

	// If status needs to be updated, also look for the ConfigMap
	if needsUpdate {
		// Look for ConfigMap with the appropriate label
		configMapList := &corev1.ConfigMapList{}
		err := r.List(ctx, configMapList, client.InNamespace(mcpServerImportJob.Namespace), client.MatchingLabels(map[string]string{
			McpServerImportJobLabel: mcpServerImportJob.Name,
		}))
		if err != nil {
			log.Error(err, "Failed to list ConfigMaps")
			return fmt.Errorf("failed to list ConfigMaps: %w", err)
		}

		configMapName := ""
		if len(configMapList.Items) > 0 {
			configMap := &configMapList.Items[0]
			configMapName = configMap.Name
			log.Info("Found ConfigMap for McpServerImportJob", "configMapName", configMapName)

			// Set ownership of the ConfigMap to the McpServerImportJob
			if err := ctrl.SetControllerReference(mcpServerImportJob, configMap, r.Scheme); err != nil {
				log.Error(err, "Failed to set controller reference on ConfigMap", "configMapName", configMapName)
				return fmt.Errorf("failed to set controller reference on ConfigMap: %w", err)
			}

			// Update the ConfigMap to set the ownership
			if err := r.Update(ctx, configMap); err != nil {
				log.Error(err, "Failed to update ConfigMap with ownership", "configMapName", configMapName)
				return fmt.Errorf("failed to update ConfigMap with ownership: %w", err)
			}

			log.Info("Set ownership of ConfigMap to McpServerImportJob", "configMapName", configMapName)
		}

		// Update the McpServerImportJob status
		mcpServerImportJob.Status.Status = newStatus
		mcpServerImportJob.Status.ConfigMapName = configMapName

		err = r.Status().Update(ctx, mcpServerImportJob)
		if err != nil {
			log.Error(err, "Failed to update McpServerImportJob status")
			return fmt.Errorf("failed to update McpServerImportJob status: %w", err)
		}

		log.Info("Updated McpServerImportJob status",
			"name", mcpServerImportJob.Name,
			"status", newStatus,
			"configMapName", configMapName)
	}

	return nil
}

// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpserverimportjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpserverimportjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpserverimportjobs/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the McpServerImportJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *McpServerImportJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the McpServerImportJob instance
	mcpServerImportJob := &mcpv1alpha1.McpServerImportJob{}
	err := r.Get(ctx, req.NamespacedName, mcpServerImportJob)
	if err != nil {
		// Handle the case where the resource is not found
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Initialize the job and check if it needs to be started
	needsStart, err := r.initializeJob(ctx, mcpServerImportJob)
	if err != nil {
		return ctrl.Result{}, err
	}

	// If the job needs to be started, handle the job execution logic here
	if needsStart {
		log.Info("Starting McpServerImportJob",
			"name", mcpServerImportJob.Name,
			"namespace", mcpServerImportJob.Namespace)

		if err := r.startImportJob(ctx, mcpServerImportJob); err != nil {
			log.Error(err, "Failed to start import job")
			return ctrl.Result{}, err
		} else if mcpServerImportJob.Status.Status == mcpv1alpha1.ImportJobRunning {
			log.Info("Import job is already running",
				"name", mcpServerImportJob.Name,
				"namespace", mcpServerImportJob.Namespace)

			return ctrl.Result{}, nil
		}
	}

	// Check Job status
	if err := r.checkJobStatus(ctx, mcpServerImportJob); err != nil {
		log.Error(err, "Failed to check Job status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *McpServerImportJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcpv1alpha1.McpServerImportJob{}).
		Owns(&batchv1.Job{}).
		Named("mcpserverimportjob").
		Complete(r)
}
