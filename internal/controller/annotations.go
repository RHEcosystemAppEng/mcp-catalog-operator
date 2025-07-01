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

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
)

const (
	McpCatalogNameAnnotation      = "mcp.opendatahub.io/mcpcatalog"
	McpCatalogNamespaceAnnotation = "mcp.opendatahub.io/mcpcatalog-namespace"

	McpRegistryNameAnnotation      = "mcp.opendatahub.io/mcpregistry"
	McpRegistryNamespaceAnnotation = "mcp.opendatahub.io/mcpregistry-namespace"

	McpServerNameAnnotation      = "mcp.opendatahub.io/mcpserver"
	McpServerNamespaceAnnotation = "mcp.opendatahub.io/mcpserver-namespace"
	McpServerVersionAnnotation   = "mcp.opendatahub.io/mcpserver-version"
	McpServerPackageAnnotation   = "mcp.opendatahub.io/mcpserver-package"

	ServerDescriptionAnnotation     = "mcp.opendatahub.io/description"
	ServerLongDescriptionAnnotation = "mcp.opendatahub.io/long-description"
	ServerProviderAnnotation        = "mcp.opendatahub.io/server-provider"
	ServerImageProviderAnnotation   = "mcp.opendatahub.io/image-provider"
	ServerHomePageAnnotation        = "mcp.opendatahub.io/homepage"
	ServerLicenseAnnotation         = "mcp.opendatahub.io/license"
	ServerCompetenciesAnnotation    = "mcp.opendatahub.io/competencies"
)

// GetMcpCatalogFromAnnotations retrieves an McpCatalog instance using annotations
// from the given object's metadata. If the namespace annotation is not present,
// it will look in the same namespace as the object.
func GetMcpCatalogFromAnnotations(ctx context.Context, c client.Client, obj client.Object) (*mcpv1alpha1.McpCatalog, error) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		return nil, fmt.Errorf("no annotations found on object")
	}

	catalogName, exists := annotations[McpCatalogNameAnnotation]
	if !exists || catalogName == "" {
		return nil, fmt.Errorf("annotation %s is required and cannot be empty", McpCatalogNameAnnotation)
	}

	catalogNamespace := obj.GetNamespace()
	if nsAnnotation, exists := annotations[McpCatalogNamespaceAnnotation]; exists && nsAnnotation != "" {
		catalogNamespace = nsAnnotation
	}

	var mcpCatalog mcpv1alpha1.McpCatalog
	if err := c.Get(ctx, types.NamespacedName{
		Name:      catalogName,
		Namespace: catalogNamespace,
	}, &mcpCatalog); err != nil {
		return nil, fmt.Errorf("failed to get McpCatalog %s in namespace %s: %w", catalogName, catalogNamespace, err)
	}

	return &mcpCatalog, nil
}

// SetMcpCatalogAnnotations sets the catalog name and namespace annotations on an object
func SetMcpCatalogAnnotations(obj client.Object, catalogName, catalogNamespace string) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[McpCatalogNameAnnotation] = catalogName
	annotations[McpCatalogNamespaceAnnotation] = catalogNamespace

	obj.SetAnnotations(annotations)
}

// GetMcpCatalogNameFromAnnotations retrieves the catalog name from annotations
func GetMcpCatalogNameFromAnnotations(obj client.Object) (string, error) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		return "", fmt.Errorf("no annotations found on object")
	}

	catalogName, exists := annotations[McpCatalogNameAnnotation]
	if !exists || catalogName == "" {
		return "", fmt.Errorf("annotation %s is required and cannot be empty", McpCatalogNameAnnotation)
	}

	return catalogName, nil
}

// GetMcpCatalogNamespaceFromAnnotations retrieves the catalog namespace from annotations
// Returns the object's namespace if no namespace annotation is present
func GetMcpCatalogNamespaceFromAnnotations(obj client.Object) string {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		return obj.GetNamespace()
	}

	if nsAnnotation, exists := annotations[McpCatalogNamespaceAnnotation]; exists && nsAnnotation != "" {
		return nsAnnotation
	}

	return obj.GetNamespace()
}
