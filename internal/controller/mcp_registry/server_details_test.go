package mcp_registry

import (
	"testing"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/mcp-registry-operator/internal/types"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestExtractServerDetailFromCertifiedServer(t *testing.T) {
	// Create a sample McpCertifiedServer with the server-detail annotation in JSON format
	sampleServerDetail := `  {
    "id": "6a4d324e-7d12-495c-b72c-fe5803363af1",
    "name": "io.github.flux159/mcp-server-kubernetes",
    "description": "MCP Server for kubernetes management commands",
    "repository": {
      "url": "https://github.com/Flux159/mcp-server-kubernetes",
      "source": "github",
      "id": "900130551"
    },
    "version_detail": {
      "version": "0.0.1-seed",
      "release_date": "2025-05-16T19:09:01Z",
      "is_latest": true
    },
    "packages": [
      {
        "registry_name": "pypi",
        "name": "mcp-server-kubernetes",
        "version": "1.6.2"
      }
    ]
  }`

	mcpCertifiedServer := &mcpv1alpha1.McpCertifiedServer{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mcp-server-kubernetes",
			Annotations: map[string]string{
				types.ServerDetailAnnotation: sampleServerDetail,
			},
		},
		Spec: mcpv1alpha1.McpCertifiedServerSpec{
			McpServer: mcpv1alpha1.McpCertifiedServerServerSpec{
				Type: mcpv1alpha1.ServerTypeContainer,
				Container: &mcpv1alpha1.ContainerServerSpec{
					Image:   "flux159/mcp-server-kubernetes:latest",
					Command: "mcp-server-kubernetes",
					Args:    []string{"--help"},
					EnvVars: []string{"MCP_SERVER_KUBERNETES_ENABLED=true"},
				},
			},
		},
	}

	mcpServerRun := &mcpv1alpha1.McpServerRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-server",
		},
	}

	// Test the extraction
	serverDetail, err := extractServerDetailsFromCertifiedServer(mcpServerRun, mcpCertifiedServer)
	if err != nil {
		t.Fatalf("Failed to extract server detail: %v", err)
	}

	// Verify the extracted data
	if serverDetail.ID != "6a4d324e-7d12-495c-b72c-fe5803363af1" {
		t.Errorf("Expected ID '6a4d324e-7d12-495c-b72c-fe5803363af1', got '%s'", serverDetail.ID)
	}

	if serverDetail.Name != "io.github.flux159/mcp-server-kubernetes" {
		t.Errorf("Expected Name 'io.github.flux159/mcp-server-kubernetes', got '%s'", serverDetail.Name)
	}

	if serverDetail.Description != "MCP Server for kubernetes management commands" {
		t.Errorf("Expected Description 'MCP Server for kubernetes management commands', got '%s'", serverDetail.Description)
	}

	if serverDetail.Repository.URL != "https://github.com/Flux159/mcp-server-kubernetes" {
		t.Errorf("Expected Repository URL 'https://github.com/Flux159/mcp-server-kubernetes', got '%s'", serverDetail.Repository.URL)
	}

	if serverDetail.VersionDetail.Version != "0.0.1-seed" {
		t.Errorf("Expected Version '0.0.1-seed', got '%s'", serverDetail.VersionDetail.Version)
	}

	if len(serverDetail.Packages) != 1 {
		t.Errorf("Expected 1 package, got %d", len(serverDetail.Packages))
	}

	if serverDetail.Packages[0].RegistryName != "docker" {
		t.Errorf("Expected package name 'docker', got '%s'", serverDetail.Packages[0].RegistryName)
	}

	if serverDetail.Packages[0].Name != "flux159/mcp-server-kubernetes" {
		t.Errorf("Expected package name 'flux159/mcp-server-kubernetes', got '%s'", serverDetail.Packages[0].Name)
	}

	if serverDetail.Packages[0].Version != "latest" {
		t.Errorf("Expected package version 'latest', got '%s'", serverDetail.Packages[0].Version)
	}

	if len(serverDetail.Remotes) != 1 {
		t.Errorf("Expected 1 remote, got %d", len(serverDetail.Remotes))
	}

	if serverDetail.Remotes[0].TransportType != "sse" {
		t.Errorf("Expected package name 'sse', got '%s'", serverDetail.Remotes[0].TransportType)
	}

	if serverDetail.Remotes[0].URL != "http://test-server-svc:8000" {
		t.Errorf("Expected package name 'http://test-server-svc:8000', got '%s'", serverDetail.Remotes[0].URL)
	}

	if len(serverDetail.Remotes[0].Headers) != 0 {
		t.Errorf("Expected 0 headers, got %d", len(serverDetail.Remotes[0].Headers))
	}
}

func TestExtractServerDetailFromCertifiedServer_MissingAnnotation(t *testing.T) {
	mcpCertifiedServer := &mcpv1alpha1.McpCertifiedServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-server",
			Annotations: map[string]string{},
		},
	}

	mcpServerRun := &mcpv1alpha1.McpServerRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-server",
		},
	}

	_, err := extractServerDetailsFromCertifiedServer(mcpServerRun, mcpCertifiedServer)
	if err == nil {
		t.Error("Expected error when server-detail annotation is missing")
	}
}

func TestExtractServerDetailFromCertifiedServer_InvalidJSON(t *testing.T) {
	mcpCertifiedServer := &mcpv1alpha1.McpCertifiedServer{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-server",
			Annotations: map[string]string{
				types.ServerDetailAnnotation: "invalid json: [",
			},
		},
	}

	mcpServerRun := &mcpv1alpha1.McpServerRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-server",
		},
	}

	_, err := extractServerDetailsFromCertifiedServer(mcpServerRun, mcpCertifiedServer)
	if err == nil {
		t.Error("Expected error when server-detail annotation contains invalid YAML")
	}
}

func TestExtractServerDetailFromContainer(t *testing.T) {
	mcpServerRun := &mcpv1alpha1.McpServerRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: "containerized-server",
		},
		Spec: mcpv1alpha1.McpServerRunSpec{
			ServerMode: "container",
			McpServer: mcpv1alpha1.McpServerConfig{
				ServerImage: "quay.io/org/image:1.0",
			},
		},
	}

	// Test the extraction
	serverDetail := extractServerDetailsFromContainer(mcpServerRun)

	// Verify the extracted data
	_, err := uuid.Parse(serverDetail.ID)
	if err != nil {
		t.Errorf("Expected ID in UUID format, got '%s'", serverDetail.ID)
	}

	if serverDetail.Name != "containerized-server" {
		t.Errorf("Expected Name 'containerized-server', got '%s'", serverDetail.Name)
	}

	if serverDetail.Description != "" {
		t.Errorf("Expected Description '', got '%s'", serverDetail.Description)
	}

	if serverDetail.Repository.URL != types.DefaultRepositoryUrl {
		t.Errorf("Expected Repository URL '%s', got '%s'", types.DefaultRepositoryUrl, serverDetail.Repository.URL)
	}

	if serverDetail.VersionDetail.Version != "1.0" {
		t.Errorf("Expected Version '1.0', got '%s'", serverDetail.VersionDetail.Version)
	}

	if len(serverDetail.Packages) != 1 {
		t.Errorf("Expected 1 package, got %d", len(serverDetail.Packages))
	}

	if serverDetail.Packages[0].RegistryName != "docker" {
		t.Errorf("Expected package name 'docker', got '%s'", serverDetail.Packages[0].RegistryName)
	}

	if serverDetail.Packages[0].Name != "quay.io/org/image" {
		t.Errorf("Expected package name 'quay.io/org/image', got '%s'", serverDetail.Packages[0].Name)
	}

	if serverDetail.Packages[0].Version != "1.0" {
		t.Errorf("Expected package version 'latest', got '%s'", serverDetail.Packages[0].Version)
	}

	if len(serverDetail.Remotes) != 1 {
		t.Errorf("Expected 1 remote, got %d", len(serverDetail.Remotes))
	}

	if serverDetail.Remotes[0].TransportType != "sse" {
		t.Errorf("Expected package name 'sse', got '%s'", serverDetail.Remotes[0].TransportType)
	}

	if serverDetail.Remotes[0].URL != "http://containerized-server-svc:8000" {
		t.Errorf("Expected package name 'http://containerized-server-svc:8000', got '%s'", serverDetail.Remotes[0].URL)
	}

	if len(serverDetail.Remotes[0].Headers) != 0 {
		t.Errorf("Expected 0 headers, got %d", len(serverDetail.Remotes[0].Headers))
	}
}
