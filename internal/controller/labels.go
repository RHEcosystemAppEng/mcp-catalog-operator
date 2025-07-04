package controller

import (
	"context"
	"fmt"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	McpStagingAreaLabel     = "mcp.opendatahub.io/mcpstagingarea"
	McpCatalogLabel         = "mcp.opendatahub.io/mcpcatalog"
	McpRegistryLabel        = "mcp.opendatahub.io/mcpregistry"
	McpServerNameLabel      = "mcp.opendatahub.io/mcpserver"
	McpServerImportJobLabel = "mcp.opendatahub.io/mcpserverimportjob"
)

// GetMcpCatalogFromLabels retrieves an McpCatalog instance using labels
// from the given object's metadata.
// It will look in the same namespace as the object.
func GetMcpCatalogFromLabels(ctx context.Context, c client.Client, obj client.Object) (*mcpv1alpha1.McpCatalog, error) {
	catalog := &mcpv1alpha1.McpCatalog{}
	err := GetObjectFromLabels(ctx, c, obj, McpCatalogLabel, catalog)
	return catalog, err
}

// GetMcpStagingAreaFromLabels retrieves an McpStagingArea instance using labels
// from the given object's metadata.
// It will look in the same namespace as the object.
func GetMcpStagingAreaFromLabels(ctx context.Context, c client.Client, obj client.Object) (*mcpv1alpha1.McpStagingArea, error) {
	stagingArea := &mcpv1alpha1.McpStagingArea{}
	err := GetObjectFromLabels(ctx, c, obj, McpStagingAreaLabel, stagingArea)
	return stagingArea, err
}

func GetObjectFromLabels(ctx context.Context, c client.Client, obj client.Object, label string, result client.Object) error {
	labels := obj.GetLabels()
	if labels == nil {
		return fmt.Errorf("no labels found on object")
	}

	objectName, exists := labels[label]
	if !exists || objectName == "" {
		return fmt.Errorf("label %s is required and cannot be empty", label)
	}

	objectNamespace := obj.GetNamespace()

	if err := c.Get(ctx, types.NamespacedName{
		Name:      objectName,
		Namespace: objectNamespace,
	}, result); err != nil {
		return fmt.Errorf("failed to get %s '%s' in namespace %s: %w", result.GetObjectKind().GroupVersionKind().Kind, objectName, objectNamespace, err)
	}

	log := logf.FromContext(ctx)
	log.Info("found referenced object", "kind", result.GetObjectKind().GroupVersionKind().Kind, "name", result.GetName(), "namespace", result.GetNamespace())

	return nil
}
