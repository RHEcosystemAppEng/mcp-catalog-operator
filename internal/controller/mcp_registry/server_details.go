package mcp_registry

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
	"github.com/google/uuid"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/RHEcosystemAppEng/mcp-registry-operator/internal/services"
	"github.com/RHEcosystemAppEng/mcp-registry-operator/internal/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ExtractServerDetail(ctx context.Context, c client.Client, mcpServerRun *mcpv1alpha1.McpServerRun) (*mcpv1alpha1.ServerDetail, error) {
	logger := logf.FromContext(ctx)

	var serverDetails *mcpv1alpha1.ServerDetail
	var err error
	switch mcpServerRun.Spec.ServerMode {
	case "blueprint":
		mcpCertifiedServer, err0 := services.GetMcpCertifiedServerFromRef(ctx, c, mcpServerRun, *mcpServerRun.Spec.McpServer.McpServerRef)
		if err0 != nil {
			return nil, fmt.Errorf("failed to get mcp certified server: %w", err0)
		}
		serverDetails, err = extractServerDetailsFromCertifiedServer(mcpServerRun, mcpCertifiedServer)
	case "remote":
		serverDetails = extractServerDetailsFromRemote(mcpServerRun)
	case "container":
		serverDetails = extractServerDetailsFromContainer(mcpServerRun)
	default:
		return nil, fmt.Errorf("invalid server mode: %s", mcpServerRun.Spec.ServerMode)
	}

	logger.Info("Server detail extracted", "serverMode", mcpServerRun.Spec.ServerMode, "serverDetail", serverDetails)
	return serverDetails, err
}

func extractServerDetailsFromCertifiedServer(mcpServerRun *mcpv1alpha1.McpServerRun, mcpCertifiedServer *mcpv1alpha1.McpCertifiedServer) (*mcpv1alpha1.ServerDetail, error) {
	// Get the server-detail annotation
	serverDetailAnnotation, exists := mcpCertifiedServer.Annotations[types.ServerDetailAnnotation]
	if !exists {
		return nil, fmt.Errorf("server detail annotation is required")
	}

	// Parse the YAML annotation into ServerDetail
	var serverDetail mcpv1alpha1.ServerDetail
	if err := json.Unmarshal([]byte(serverDetailAnnotation), &serverDetail); err != nil {
		return nil, fmt.Errorf("failed to parse server detail annotation: %w", err)
	}

	// Validate that required fields are present
	if serverDetail.ID == "" {
		return nil, fmt.Errorf("server detail must have an ID")
	}
	if serverDetail.Name == "" {
		return nil, fmt.Errorf("server detail must have a name")
	}
	if serverDetail.Description == "" {
		return nil, fmt.Errorf("server detail must have a description")
	}

	serverImage := mcpCertifiedServer.Spec.McpServer.Container.Image
	serverBaseImage := strings.Split(serverImage, ":")[0]
	imageTag := strings.Split(serverImage, ":")[1]
	operatorServerDetail := &mcpv1alpha1.ServerDetail{
		Server: mcpv1alpha1.Server{
			ID:            serverDetail.ID,
			Name:          serverDetail.Name,
			Description:   serverDetail.Description,
			Repository:    serverDetail.Repository,
			VersionDetail: serverDetail.VersionDetail,
		},
		Packages: []mcpv1alpha1.Package{
			{
				RegistryName: "docker",
				Name:         serverBaseImage,
				Version:      imageTag,
			},
		},
		Remotes: []mcpv1alpha1.Remote{
			{
				TransportType: "sse",                                       // TODO
				URL:           "http://" + mcpServerRun.Name + "-svc:8000", // TODO make this 80000 configurable
				Headers:       []mcpv1alpha1.Input{},
			},
		},
	}

	return operatorServerDetail, nil
}

func extractServerDetailsFromRemote(mcpServerRun *mcpv1alpha1.McpServerRun) *mcpv1alpha1.ServerDetail {
	operatorServerDetail := &mcpv1alpha1.ServerDetail{
		Server: mcpv1alpha1.Server{
			ID:          uuid.New().String(), // TODO store in McpServerRun instead?
			Name:        mcpServerRun.Name,
			Description: "", // TODO mcpServerRun.Description,
			Repository: mcpv1alpha1.Repository{
				URL:    types.DefaultRepositoryUrl, // TODO
				Source: "github",
				ID:     "NA",
			},
			VersionDetail: mcpv1alpha1.VersionDetail{
				Version:     "NA",
				ReleaseDate: time.Now().Format(time.RFC3339), // TODO
				IsLatest:    true,                            // TODO
			},
		},
		Packages: []mcpv1alpha1.Package{},
		Remotes: []mcpv1alpha1.Remote{
			{
				TransportType: "sse",                                       // TODO
				URL:           "http://" + mcpServerRun.Name + "-svc:8000", // TODO make this 80000 configurable
				Headers:       []mcpv1alpha1.Input{},
			},
		},
	}

	return operatorServerDetail
}

func extractServerDetailsFromContainer(mcpServerRun *mcpv1alpha1.McpServerRun) *mcpv1alpha1.ServerDetail {
	serverImage := mcpServerRun.Spec.McpServer.ServerImage
	var serverBaseImage string
	var imageTag string
	if strings.Contains(serverImage, ":") {
		serverBaseImage = strings.Split(serverImage, ":")[0]
		imageTag = strings.Split(serverImage, ":")[1]
	} else {
		serverBaseImage = serverImage
		imageTag = "latest"
	}

	operatorServerDetail := &mcpv1alpha1.ServerDetail{
		Server: mcpv1alpha1.Server{
			ID:          uuid.New().String(), // TODO store in McpServerRun instead?
			Name:        mcpServerRun.Name,
			Description: "", // TODO mcpServerRun.Description,
			Repository: mcpv1alpha1.Repository{
				URL:    types.DefaultRepositoryUrl, // TODO
				Source: "github",
				ID:     "NA",
			},
			VersionDetail: mcpv1alpha1.VersionDetail{
				Version:     imageTag,
				ReleaseDate: time.Now().Format(time.RFC3339), // TODO
				IsLatest:    true,                            // TODO
			},
		},
		Packages: []mcpv1alpha1.Package{
			{
				RegistryName: "docker",
				Name:         serverBaseImage,
				Version:      imageTag,
			},
		},
		Remotes: []mcpv1alpha1.Remote{
			{
				TransportType: "sse",                                       // TODO
				URL:           "http://" + mcpServerRun.Name + "-svc:8000", // TODO make this 80000 configurable
				Headers:       []mcpv1alpha1.Input{},
			},
		},
	}

	return operatorServerDetail
}
