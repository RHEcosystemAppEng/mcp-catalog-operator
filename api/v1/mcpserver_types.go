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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
type ServerRef struct {
	Name      string  `json:"name"`
	Namespace *string `json:"namespace,omitempty"`
}

// START copy from https://github.com/modelcontextprotocol/registry

// AuthMethod represents the authentication method used
type AuthMethod string

const (
	// AuthMethodGitHub represents GitHub OAuth authentication
	AuthMethodGitHub AuthMethod = "github"
	// AuthMethodNone represents no authentication
	AuthMethodNone AuthMethod = "none"
)

// Authentication holds information about the authentication method and credentials
type Authentication struct {
	Method  AuthMethod `json:"method,omitempty"`
	Token   string     `json:"token,omitempty"`
	RepoRef string     `json:"repo_ref,omitempty"`
}

// Repository represents a source code repository as defined in the spec
type Repository struct {
	URL    string `json:"url"`
	Source string `json:"source"`
	ID     string `json:"id"`
}

// create an enum for Format
type Format string

const (
	FormatString   Format = "string"
	FormatNumber   Format = "number"
	FormatBoolean  Format = "boolean"
	FormatFilePath Format = "file_path"
)

// UserInput represents a user input as defined in the spec
type Input struct {
	Description string            `json:"description,omitempty"`
	IsRequired  bool              `json:"is_required,omitempty"`
	Format      Format            `json:"format,omitempty"`
	Value       string            `json:"value,omitempty"`
	IsSecret    bool              `json:"is_secret,omitempty"`
	Default     string            `json:"default,omitempty"`
	Choices     []string          `json:"choices,omitempty"`
	Template    string            `json:"template,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
}

type InputWithVariables struct {
	Input     `json:",inline"`
	Variables map[string]Input `json:"variables,omitempty"`
}

type KeyValueInput struct {
	InputWithVariables `json:",inline"`
	Name               string `json:"name"`
}
type ArgumentType string

const (
	ArgumentTypePositional ArgumentType = "positional"
	ArgumentTypeNamed      ArgumentType = "named"
)

// RuntimeArgument defines a type that can be either a PositionalArgument or a NamedArgument
type Argument struct {
	InputWithVariables `json:",inline"`
	Type               ArgumentType `json:"type"`
	Name               string       `json:"name,omitempty"`
	IsRepeated         bool         `json:"is_repeated,omitempty"`
	ValueHint          string       `json:"value_hint,omitempty"`
}

type Package struct {
	RegistryName         string          `json:"registry_name"`
	Name                 string          `json:"name"`
	Version              string          `json:"version"`
	RunTimeHint          string          `json:"runtime_hint,omitempty"`
	RuntimeArguments     []Argument      `json:"runtime_arguments,omitempty"`
	PackageArguments     []Argument      `json:"package_arguments,omitempty"`
	EnvironmentVariables []KeyValueInput `json:"environment_variables,omitempty"`
}

// Remote represents a remote connection endpoint
type Remote struct {
	TransportType string  `json:"transport_type"`
	URL           string  `json:"url"`
	Headers       []Input `json:"headers,omitempty"`
}

// VersionDetail represents the version details of a server
type VersionDetail struct {
	Version     string `json:"version"`
	ReleaseDate string `json:"release_date"`
	IsLatest    bool   `json:"is_latest"`
}

// Server represents a basic server information as defined in the spec
type Server struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Repository    Repository    `json:"repository"`
	VersionDetail VersionDetail `json:"version_detail"`
}

// ServerDetail represents detailed server information as defined in the spec
type ServerDetail struct {
	Server   `json:",inline"`
	Packages []Package `json:"packages,omitempty"`
	Remotes  []Remote  `json:"remotes,omitempty"`
}

// END copy from https://github.com/modelcontextprotocol/registry

// McpServerSpec defines the desired state of McpServer.
type McpServerSpec struct {
	CatalogRef   CatalogRef   `json:"catalogRef,omitempty"`
	ServerDetail ServerDetail `json:"server_detail"`
}

// McpServerStatus defines the observed state of McpServer.
type McpServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// McpServer is the Schema for the mcpservers API.
type McpServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   McpServerSpec   `json:"spec,omitempty"`
	Status McpServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// McpServerList contains a list of McpServer.
type McpServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []McpServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&McpServer{}, &McpServerList{})
}
