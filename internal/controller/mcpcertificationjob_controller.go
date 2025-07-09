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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/mcp-registry-operator/internal/services"
	"github.com/RHEcosystemAppEng/mcp-registry-operator/internal/types"
)

// McpCertificationJobReconciler reconciles a McpCertificationJob object
type McpCertificationJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpcertificationjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpcertificationjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpcertificationjobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the McpCertificationJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *McpCertificationJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	mcpCertificationJob := &mcpv1alpha1.McpCertificationJob{}
	err := r.Get(ctx, req.NamespacedName, mcpCertificationJob)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err = r.initializeOwnershipAndReadiness(ctx, mcpCertificationJob, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// initializeOwnershipAndReadiness sets the owner reference to the referenced McpCatalog and sets the Ready condition accordingly.
func (r *McpCertificationJobReconciler) initializeOwnershipAndReadiness(ctx context.Context, mcpCertificationJob *mcpv1alpha1.McpCertificationJob, log logr.Logger) error {
	// Get McpCatalog using annotations
	mcpCatalog, err := services.GetMcpCatalogFromLabels(ctx, r.Client, mcpCertificationJob)
	catalogExists := err == nil

	var readyCondition metav1.Condition
	now := metav1.NewTime(time.Now())
	if catalogExists {
		if mcpCatalog.Namespace != mcpCertificationJob.Namespace {
			readyCondition = metav1.Condition{
				Type:               types.ConditionTypeReady,
				Status:             metav1.ConditionFalse,
				Reason:             types.ConditionReasonCrossNamespaces,
				Message:            types.ValidationMessageCrossNamespaces,
				LastTransitionTime: now,
				ObservedGeneration: mcpCertificationJob.Generation,
			}
		} else {
			readyCondition = metav1.Condition{
				Type:               types.ConditionTypeReady,
				Status:             metav1.ConditionTrue,
				Reason:             types.ConditionReasonValidationSucceeded,
				Message:            "McpCertificationJob spec is valid",
				LastTransitionTime: now,
				ObservedGeneration: mcpCertificationJob.Generation,
			}
			if err := controllerutil.SetControllerReference(mcpCatalog, mcpCertificationJob, r.Scheme); err != nil {
				return fmt.Errorf("failed to set owner reference: %w", err)
			}
			if err := r.Update(ctx, mcpCertificationJob); err != nil {
				return fmt.Errorf("failed to update McpCertificationJob with owner ref: %w", err)
			}
		}
	} else {
		readyCondition = metav1.Condition{
			Type:               types.ConditionTypeReady,
			Status:             metav1.ConditionFalse,
			Reason:             types.ConditionReasonValidationFailed,
			Message:            types.ValidationMessageCatalogNotFound,
			LastTransitionTime: now,
			ObservedGeneration: mcpCertificationJob.Generation,
		}
		log.Info("Referenced catalog NOT found", "error", err)
	}

	meta.SetStatusCondition(&mcpCertificationJob.Status.Conditions, readyCondition)
	if err := r.Status().Update(ctx, mcpCertificationJob); err != nil {
		log.Error(err, "Failed to update McpCertificationJob status")
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *McpCertificationJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcpv1alpha1.McpCertificationJob{}).
		Named("mcpcertificationjob").
		Complete(r)
}
