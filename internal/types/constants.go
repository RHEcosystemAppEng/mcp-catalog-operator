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

package types

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

	// Registry constants
	DefaultRepositoryUrl = "https://github.com/modelcontextprotocol/modelcontextprotocol"
	RegistryImage        = "quay.io/ecosystem-appeng/mcp-registry@sha256:244dc363914a30a9a24fd7ea6a0e709bb885b806ca401354450b479e13e1a16b"
	RegistryPort         = 8080
	RegistryUiImage      = "quay.io/maorfr/mcp-registry-ui:latest"
	RegistryUiPort       = 8080

	// MongoDB constants
	MongoDBDatabaseType   = "mongodb"
	MongoDBDatabaseName   = "mcp-registry"
	MongoDBCollectionName = "servers_v2"
	MongoDBLogLevel       = "debug"
	MongoDBSeedImport     = "false"
	MongoDBPort           = "27017"
)
