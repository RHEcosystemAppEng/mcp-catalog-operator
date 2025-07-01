package controller

import (
	"context"
	"fmt"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	McpCatalogNameLabel     = "mcp.opendatahub.io/mcpcatalog"
	McpServerImportJobLabel = "mcp.opendatahub.io/mcpserverimportjob"
)

// GetMcpCatalogFromLabels retrieves an McpCatalog instance using annotations
// from the given object's metadata. If the namespace annotation is not present,
// it will look in the same namespace as the object.
func GetMcpCatalogFromLabels(ctx context.Context, c client.Client, obj client.Object) (*mcpv1alpha1.McpCatalog, error) {
	labels := obj.GetLabels()
	if labels == nil {
		return nil, fmt.Errorf("no labels found on object")
	}

	catalogName, exists := labels[McpCatalogNameLabel]
	if !exists || catalogName == "" {
		return nil, fmt.Errorf("label %s is required and cannot be empty", McpCatalogNameLabel)
	}

	catalogNamespace := obj.GetNamespace()

	var mcpCatalog mcpv1alpha1.McpCatalog
	if err := c.Get(ctx, types.NamespacedName{
		Name:      catalogName,
		Namespace: catalogNamespace,
	}, &mcpCatalog); err != nil {
		return nil, fmt.Errorf("failed to get McpCatalog %s in namespace %s: %w", catalogName, catalogNamespace, err)
	}

	return &mcpCatalog, nil
}
