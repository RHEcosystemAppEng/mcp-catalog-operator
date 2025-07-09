package services

import (
	"context"
	"fmt"

	"github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func GetMcpCertifiedServerFromRef(ctx context.Context, c client.Client, obj client.Object, serverRef v1alpha1.McpServerRef) (*v1alpha1.McpCertifiedServer, error) {
	logger := log.FromContext(ctx)

	if serverRef.Name == "" {
		return nil, fmt.Errorf("invalid mcpServerRef: name missing")
	}

	mcpCertServer := &v1alpha1.McpCertifiedServer{}
	ns := obj.GetNamespace()
	if serverRef.Namespace != nil {
		ns = *serverRef.Namespace
	}
	logger.Info("Looking for McpCertifiedServer", "name", serverRef.Name, "namespace", ns)
	if err := c.Get(ctx, k8stypes.NamespacedName{Name: serverRef.Name, Namespace: ns}, mcpCertServer); err != nil {
		return nil, fmt.Errorf("failed to get referenced McpCertifiedServer: %w", err)
	}

	return mcpCertServer, nil
}
