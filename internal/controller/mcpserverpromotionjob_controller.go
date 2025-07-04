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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
)

// McpServerPromotionJobReconciler reconciles a McpServerPromotionJob object
type McpServerPromotionJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpserverpromotionjob,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpserverpromotionjob/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpserverpromotionjob/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the McpServerPromotionJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *McpServerPromotionJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the McpServerPromotionJob instance
	promotionJob := &mcpv1alpha1.McpServerPromotionJob{}
	if err := r.Get(ctx, req.NamespacedName, promotionJob); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get the referenced McpStagingArea using labels
	stagingArea, err := GetMcpStagingAreaFromLabels(ctx, r.Client, promotionJob)
	if err != nil {
		log.Error(err, "Failed to get referenced McpStagingArea")
		return ctrl.Result{}, err
	}

	// Set the owner reference to the staging area
	if err := controllerutil.SetControllerReference(stagingArea, promotionJob, r.Scheme); err != nil {
		log.Error(err, "Failed to set owner reference to McpStagingArea")
		return ctrl.Result{}, err
	}
	if err := r.Update(ctx, promotionJob); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update McpServerPromotionJob with owner ref: %w", err)
	}

	// Initialize ServerPromotions if not already set or if the list is empty
	if len(promotionJob.Status.ServerPromotions) == 0 {
		serverPromotions := make([]mcpv1alpha1.ServerPromotionStatusDefinition, 0, len(promotionJob.Spec.Servers))
		for _, serverName := range promotionJob.Spec.Servers {
			serverPromotions = append(serverPromotions, mcpv1alpha1.ServerPromotionStatusDefinition{
				Name:             serverName,
				PromotionStatus:  mcpv1alpha1.PromotionStatusPlanned,
				OriginalImage:    "",
				DestinationImage: "",
			})
		}
		promotionJob.Status.ServerPromotions = serverPromotions
		if err := r.Status().Update(ctx, promotionJob); err != nil {
			log.Error(err, "Failed to update McpServerPromotionJob status with initial server promotions")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *McpServerPromotionJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcpv1alpha1.McpServerPromotionJob{}).
		Named("mcpserverpromotionjob").
		Complete(r)
}
