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
	k8stypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/mcp-registry-operator/internal/types"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// McpImportJobReconciler reconciles a McpImportJob object
type McpImportJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// initializeJob initializes missing fields in the McpImportJob
// Returns an error if initialization fails, and a boolean indicating if the job needs to be started
func (r *McpImportJobReconciler) initializeJob(ctx context.Context, mcpImportJob *mcpv1alpha1.McpImportJob) (bool, error) {
	log := logf.FromContext(ctx)

	// Initialize missing fields in spec
	needsUpdate := false
	needsStatusUpdate := false
	needsStart := false
	defaultNameFilter := ""
	defaultMaxServers := types.DefaultMaxServers
	defaultConfigMapName := ""

	// Initialize NameFilter to empty string if not set
	if mcpImportJob.Spec.NameFilter == nil {
		mcpImportJob.Spec.NameFilter = &defaultNameFilter
		needsUpdate = true
	}

	// Initialize MaxServers to 10 if not set
	if mcpImportJob.Spec.MaxServers == nil {
		mcpImportJob.Spec.MaxServers = &defaultMaxServers
		needsUpdate = true
	}

	// Initialize status if not set
	if mcpImportJob.Status.Status == "" {
		mcpImportJob.Status.Status = mcpv1alpha1.ImportJobRunning
		needsStatusUpdate = true
		needsStart = true
	}

	// Initialize ConfigMapName to empty string if not set
	if mcpImportJob.Status.ConfigMapName == nil {
		mcpImportJob.Status.ConfigMapName = &defaultConfigMapName
		needsStatusUpdate = true
	}

	if needsStatusUpdate {
		err := r.Status().Update(ctx, mcpImportJob)
		if err != nil {
			log.Error(err, "Failed to update McpImportJob status")
			return false, err
		}
	}

	catalogName := mcpImportJob.Labels[types.McpCatalogLabel]
	if catalogName == "" {
		return false, fmt.Errorf("McpCatalogLabel not found on McpImportJob")
	}

	mcpCatalog := &mcpv1alpha1.McpCatalog{}
	err := r.Get(ctx, k8stypes.NamespacedName{Name: catalogName, Namespace: mcpImportJob.Namespace}, mcpCatalog)
	if err != nil {
		return false, fmt.Errorf("failed to get McpCatalog: %w", err)
	}
	if err := controllerutil.SetControllerReference(mcpCatalog, mcpImportJob, r.Scheme); err != nil {
		return false, fmt.Errorf("failed to set owner reference: %w", err)
	}
	log.Info("Set ownership of McpImportJob to McpCatalog", "catalogName", catalogName)

	// Update the resource if any fields were initialized
	if needsUpdate {
		log.Info("Initializing McpImportJob fields",
			"name", mcpImportJob.Name,
			"namespace", mcpImportJob.Namespace)

		err := r.Update(ctx, mcpImportJob)
		if err != nil {
			log.Error(err, "Failed to update McpImportJob spec")
			return false, err
		}
	}

	return needsStart, nil
}

// createServiceAccount creates the mcpserver-importer ServiceAccount if it doesn't exist
func (r *McpImportJobReconciler) createServiceAccount(ctx context.Context, namespace string) error {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      types.McpServerImporterServiceAccountName,
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
func (r *McpImportJobReconciler) createRole(ctx context.Context, namespace string) error {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      types.McpServerImporterRoleName,
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
func (r *McpImportJobReconciler) createRoleBinding(ctx context.Context, namespace string) error {
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      types.McpServerImporterRoleBindingName,
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      types.McpServerImporterServiceAccountName,
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     types.McpServerImporterRoleName,
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
func (r *McpImportJobReconciler) createImportJob(ctx context.Context, mcpImportJob *mcpv1alpha1.McpImportJob) error {
	log := logf.FromContext(ctx)

	// Get the catalog name from labels
	catalogName := mcpImportJob.Labels[types.McpCatalogLabel]
	if catalogName == "" {
		return fmt.Errorf("McpCatalogLabel not found on McpImportJob")
	}

	// Convert MaxServers to string
	maxServersStr := strconv.Itoa(types.DefaultMaxServers) // default value
	if mcpImportJob.Spec.MaxServers != nil {
		maxServersStr = strconv.Itoa(*mcpImportJob.Spec.MaxServers)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: types.McpServerImporterJobGenerateName,
			Namespace:    mcpImportJob.Namespace,
			Labels: map[string]string{
				types.McpCatalogLabel:   catalogName,
				types.McpImportJobLabel: mcpImportJob.Name,
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: types.McpServerImporterServiceAccountName,
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name:  types.McpServerImporterContainerName,
							Image: types.McpServerImporterImage,
							Env: []corev1.EnvVar{
								{
									Name:  "CATALOG_NAME",
									Value: catalogName,
								},
								{
									Name:  "REGISTRY_URL",
									Value: mcpImportJob.Spec.RegistryURI,
								},
								{
									Name:  "IMPORT_JOB_NAME",
									Value: mcpImportJob.Name,
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

	if mcpImportJob.Spec.NameFilter != nil {
		job.Spec.Template.Spec.Containers[0].Env = append(job.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  "NAME_FILTER",
			Value: *mcpImportJob.Spec.NameFilter,
		})
	}

	// Set the owner reference to make the Job owned by the McpImportJob
	if err := controllerutil.SetControllerReference(mcpImportJob, job, r.Scheme); err != nil {
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
func (r *McpImportJobReconciler) startImportJob(ctx context.Context, mcpImportJob *mcpv1alpha1.McpImportJob) error {
	log := logf.FromContext(ctx)

	// Create ServiceAccount
	if err := r.createServiceAccount(ctx, mcpImportJob.Namespace); err != nil {
		log.Error(err, "Failed to create ServiceAccount")
		return err
	}

	// Create Role
	if err := r.createRole(ctx, mcpImportJob.Namespace); err != nil {
		log.Error(err, "Failed to create Role")
		return err
	}

	// Create RoleBinding
	if err := r.createRoleBinding(ctx, mcpImportJob.Namespace); err != nil {
		log.Error(err, "Failed to create RoleBinding")
		return err
	}

	// Create the import Job
	if err := r.createImportJob(ctx, mcpImportJob); err != nil {
		log.Error(err, "Failed to create import Job")
		return err
	}

	log.Info("Successfully started import job", "name", mcpImportJob.Name, "namespace", mcpImportJob.Namespace)
	return nil
}

// checkJobStatus checks the status of the Job owned by this McpImportJob and updates the status accordingly
func (r *McpImportJobReconciler) checkJobStatus(ctx context.Context, mcpImportJob *mcpv1alpha1.McpImportJob) error {
	log := logf.FromContext(ctx)

	// List Jobs owned by this McpImportJob
	jobList := &batchv1.JobList{}
	err := r.List(ctx, jobList, client.InNamespace(mcpImportJob.Namespace), client.MatchingLabels(map[string]string{
		types.McpImportJobLabel: mcpImportJob.Name,
	}))
	if err != nil {
		return fmt.Errorf("failed to list Jobs: %w", err)
	}

	if len(jobList.Items) == 0 {
		log.Info("No Jobs found for McpImportJob", "name", mcpImportJob.Name, "namespace", mcpImportJob.Namespace)
		return nil
	}

	// Get the first Job (there should only be one)
	job := &jobList.Items[0]
	log.Info("Found Job for McpImportJob", "jobName", job.Name, "jobStatus", job.Status.Conditions)

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
		err := r.List(ctx, configMapList, client.InNamespace(mcpImportJob.Namespace), client.MatchingLabels(map[string]string{
			types.McpImportJobLabel: mcpImportJob.Name,
		}))
		if err != nil {
			log.Error(err, "Failed to list ConfigMaps")
			return fmt.Errorf("failed to list ConfigMaps: %w", err)
		}

		configMapName := ""
		if len(configMapList.Items) > 0 {
			configMap := &configMapList.Items[0]
			configMapName = configMap.Name
			log.Info("Found ConfigMap for McpImportJob", "configMapName", configMapName)

			// Set ownership of the ConfigMap to the McpImportJob
			if err := controllerutil.SetControllerReference(mcpImportJob, configMap, r.Scheme); err != nil {
				log.Error(err, "Failed to set controller reference on ConfigMap", "configMapName", configMapName)
				return fmt.Errorf("failed to set controller reference on ConfigMap: %w", err)
			}

			// Update the ConfigMap to set the ownership
			if err := r.Update(ctx, configMap); err != nil {
				log.Error(err, "Failed to update ConfigMap with ownership", "configMapName", configMapName)
				return fmt.Errorf("failed to update ConfigMap with ownership: %w", err)
			}

			log.Info("Set ownership of ConfigMap to McpImportJob", "configMapName", configMapName)
		}

		// Update the McpImportJob status
		mcpImportJob.Status.Status = newStatus
		mcpImportJob.Status.ConfigMapName = &configMapName

		err = r.Status().Update(ctx, mcpImportJob)
		if err != nil {
			log.Error(err, "Failed to update McpImportJob status")
			return fmt.Errorf("failed to update McpImportJob status: %w", err)
		}

		log.Info("Updated McpImportJob status",
			"name", mcpImportJob.Name,
			"status", newStatus,
			"configMapName", configMapName)
	}

	return nil
}

// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpimportjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpimportjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpimportjobs/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the McpImportJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *McpImportJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the McpImportJob instance
	mcpImportJob := &mcpv1alpha1.McpImportJob{}
	err := r.Get(ctx, req.NamespacedName, mcpImportJob)
	if err != nil {
		// Handle the case where the resource is not found
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Initialize the job and check if it needs to be started
	needsStart, err := r.initializeJob(ctx, mcpImportJob)
	if err != nil {
		return ctrl.Result{}, err
	}

	// If the job needs to be started, handle the job execution logic here
	if needsStart {
		log.Info("Starting McpImportJob",
			"name", mcpImportJob.Name,
			"namespace", mcpImportJob.Namespace)

		if err := r.startImportJob(ctx, mcpImportJob); err != nil {
			log.Error(err, "Failed to start import job")
			return ctrl.Result{}, err
		} else if mcpImportJob.Status.Status == mcpv1alpha1.ImportJobRunning {
			log.Info("Import job is already running",
				"name", mcpImportJob.Name,
				"namespace", mcpImportJob.Namespace)

			return ctrl.Result{}, nil
		}
	}

	// Check Job status
	if err := r.checkJobStatus(ctx, mcpImportJob); err != nil {
		log.Error(err, "Failed to check Job status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *McpImportJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcpv1alpha1.McpImportJob{}).
		Owns(&batchv1.Job{}).
		Named("mcpimportjob").
		Complete(r)
}
