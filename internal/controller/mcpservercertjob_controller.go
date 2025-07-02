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
)

// McpServerCertJobReconciler reconciles a McpServerCertJob object
type McpServerCertJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpservercertjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpservercertjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpservercertjobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the McpServerCertJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *McpServerCertJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	mcpServerCertJob := &mcpv1alpha1.McpServerCertJob{}
	err := r.Get(ctx, req.NamespacedName, mcpServerCertJob)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err = r.initializeOwnershipAndReadiness(ctx, mcpServerCertJob, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// initializeOwnershipAndReadiness sets the owner reference to the referenced McpCatalog and sets the Ready condition accordingly.
func (r *McpServerCertJobReconciler) initializeOwnershipAndReadiness(ctx context.Context, mcpServerCertJob *mcpv1alpha1.McpServerCertJob, log logr.Logger) error {
	now := metav1.NewTime(time.Now())
	mcpCatalog, err := GetMcpCatalogFromLabels(ctx, r.Client, mcpServerCertJob)
	catalogExists := err == nil

	var readyCondition metav1.Condition
	if catalogExists {
		log.Info("Referenced catalog found", "catalog", mcpCatalog.Name, "namespace", mcpCatalog.Namespace)
		if mcpCatalog.Namespace != mcpServerCertJob.Namespace {
			readyCondition = metav1.Condition{
				Type:               ConditionTypeReady,
				Status:             metav1.ConditionFalse,
				Reason:             ConditionReasonCrossNamespaces,
				Message:            ValidationMessageCrossNamespaces,
				LastTransitionTime: now,
				ObservedGeneration: mcpServerCertJob.Generation,
			}
		} else {
			readyCondition = metav1.Condition{
				Type:               ConditionTypeReady,
				Status:             metav1.ConditionTrue,
				Reason:             ConditionReasonValidationSucceeded,
				Message:            ValidationMessageServerSuccess,
				LastTransitionTime: now,
				ObservedGeneration: mcpServerCertJob.Generation,
			}
			if err := controllerutil.SetControllerReference(mcpCatalog, mcpServerCertJob, r.Scheme); err != nil {
				return fmt.Errorf("failed to set owner reference: %w", err)
			}
			if err := r.Update(ctx, mcpServerCertJob); err != nil {
				return fmt.Errorf("failed to update McpServerCertJob with owner ref: %w", err)
			}
		}
	} else {
		readyCondition = metav1.Condition{
			Type:               ConditionTypeReady,
			Status:             metav1.ConditionFalse,
			Reason:             ConditionReasonValidationFailed,
			Message:            ValidationMessageCatalogNotFound,
			LastTransitionTime: now,
			ObservedGeneration: mcpServerCertJob.Generation,
		}
		log.Info("Referenced catalog NOT found", "error", err)
	}

	meta.SetStatusCondition(&mcpServerCertJob.Status.Conditions, readyCondition)
	if err := r.Status().Update(ctx, mcpServerCertJob); err != nil {
		log.Error(err, "Failed to update McpServerCertJob status")
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *McpServerCertJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcpv1alpha1.McpServerCertJob{}).
		Named("mcpservercertjob").
		Complete(r)
}
