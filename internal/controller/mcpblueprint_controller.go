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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mcpv1 "github.com/dmartinol/mcp-catalog-operator/api/v1"
)

// McpBlueprintReconciler reconciles a McpBlueprint object
type McpBlueprintReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpblueprints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpblueprints/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpblueprints/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the McpBlueprint object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *McpBlueprintReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// TODO(user): your logic here
	var mcpBlueprint mcpv1.McpBlueprint
	if err := r.Get(ctx, req.NamespacedName, &mcpBlueprint); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	ref := mcpBlueprint.Spec.CatalogRef
	if ref.Name == "" {
		return ctrl.Result{}, fmt.Errorf("invalid catalog-ref: name missing")
	}

	var mcpCatalog mcpv1.McpCatalog
	ns := mcpBlueprint.Namespace
	if ref.Namespace != nil {
		ns = *ref.Namespace
	}
	fmt.Printf("Looking for McpCatalog %s in %s", ref.Name, ns)
	if err := r.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: ns}, &mcpCatalog); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get referenced McpCatalog: %w", err)
	}

	// Set McpCatalog as owner of McpBlueprint
	if err := controllerutil.SetControllerReference(&mcpCatalog, &mcpBlueprint, r.Scheme); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set owner reference: %w", err)
	}

	// Persist change if necessary
	if err := r.Update(ctx, &mcpBlueprint); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update McpBlueprint with owner ref: %w", err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *McpBlueprintReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcpv1.McpBlueprint{}).
		Named("mcpblueprint").
		Complete(r)
}
