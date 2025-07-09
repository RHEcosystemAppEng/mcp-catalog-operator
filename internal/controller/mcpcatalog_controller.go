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
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/mcp-registry-operator/internal/types"
)

// McpCatalogReconciler reconciles a McpCatalog object
type McpCatalogReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpcatalogs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpcatalogs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpcatalogs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the McpCatalog object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *McpCatalogReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the McpCatalog instance
	mcpCatalog := &mcpv1alpha1.McpCatalog{}
	err := r.Get(ctx, req.NamespacedName, mcpCatalog)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Validate the McpCatalog spec
	validationErr := r.validateMcpCatalog(mcpCatalog)

	// Update status based on validation result
	if validationErr != nil {
		log.Error(validationErr, "Validation failed for McpCatalog", "name", mcpCatalog.Name, "namespace", mcpCatalog.Namespace)
		r.setReadyCondition(mcpCatalog, metav1.ConditionFalse, types.ConditionReasonValidationFailed, validationErr.Error())
	} else {
		log.Info("Successfully validated McpCatalog", "name", mcpCatalog.Name, "namespace", mcpCatalog.Namespace)
		r.setReadyCondition(mcpCatalog, metav1.ConditionTrue, types.ConditionReasonValidationSucceeded, types.ValidationMessageCatalogSuccess)
	}

	// Update the status
	if err := r.Client.Status().Update(ctx, mcpCatalog); err != nil {
		log.Error(err, "Failed to update McpCatalog status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// validateMcpCatalog validates the McpCatalog spec and returns a descriptive error if validation fails
func (r *McpCatalogReconciler) validateMcpCatalog(mcpCatalog *mcpv1alpha1.McpCatalog) error {
	var validationErrors []string

	// Check if description is empty or null
	if strings.TrimSpace(mcpCatalog.Spec.Description) == "" {
		validationErrors = append(validationErrors, types.ValidationMessageDescriptionRequired)
	}

	// Check if imageRegistry is empty or null
	if strings.TrimSpace(mcpCatalog.Spec.ImageRegistry) == "" {
		validationErrors = append(validationErrors, types.ValidationMessageImageRegistryRequired)
	}

	// If there are validation errors, return a descriptive error
	if len(validationErrors) > 0 {
		return fmt.Errorf("McpCatalog validation failed for %s/%s: %s",
			mcpCatalog.Namespace, mcpCatalog.Name, strings.Join(validationErrors, "; "))
	}

	return nil
}

// setReadyCondition sets the Ready condition on the McpCatalog
func (r *McpCatalogReconciler) setReadyCondition(mcpCatalog *mcpv1alpha1.McpCatalog, status metav1.ConditionStatus, reason, message string) {
	now := metav1.NewTime(time.Now())

	readyCondition := metav1.Condition{
		Type:               types.ConditionTypeReady,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
		ObservedGeneration: mcpCatalog.Generation,
	}

	// Update or add the condition
	meta.SetStatusCondition(&mcpCatalog.Status.Conditions, readyCondition)
}

// SetupWithManager sets up the controller with the Manager.
func (r *McpCatalogReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcpv1alpha1.McpCatalog{}).
		Named("mcpcatalog").
		Complete(r)
}
