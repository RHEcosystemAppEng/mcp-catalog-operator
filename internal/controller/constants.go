package controller

const (
	// ServiceAccount and RBAC constants
	McpServerImporterServiceAccountName = "mcpserver-importer"
	McpServerImporterRoleName           = "mcpserver-importer"
	McpServerImporterRoleBindingName    = "mcpserver-importer"

	// Job constants
	McpServerImporterJobGenerateName = "mcpserver-importer-job-"
	McpServerImporterContainerName   = "mcp-server-importer"
	McpServerImporterImage           = "quay.io/ecosystem-appeng/mcpserver-importer@sha256:98dd91fd2ecdd49a225fd8518705a01bf0dc5008141562017252b431f9149037"

	// Default values
	DefaultMaxServers = 10
)
