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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=quarantined;in-review;certified;rejected
type CertificationStatus string

const (
	CertificationStatusQuarantined CertificationStatus = "quarantined"
	CertificationStatusInReview    CertificationStatus = "in-review"
	CertificationStatusCertified   CertificationStatus = "certified"
	CertificationStatusRejected    CertificationStatus = "rejected"
)

type ImageBuildStatus string

const (
	ImageBuildStatusNotStarted ImageBuildStatus = "not-started"
	ImageBuildStatusInProgress ImageBuildStatus = "in-progress"
	ImageBuildStatusCompleted  ImageBuildStatus = "completed"
	ImageBuildStatusFailed     ImageBuildStatus = "failed"
)

// +kubebuilder:validation:Enum=container;remote
type ServerType string

const (
	ServerTypeContainer ServerType = "container"
	ServerTypeRemote    ServerType = "remote"
)

// ContainerServerSpec defines the configuration for a container-based MCP server
type ContainerServerSpec struct {
	Image string `json:"image"`
	// The command to run the server
	Command string `json:"command"`
	// The arguments to pass to the server
	Args []string `json:"args"`
	// The environment variables to set for the server
	EnvVars []string `json:"envVars"`
}

// RemoteServerSpec defines the configuration for a remote MCP server
type RemoteServerSpec struct {
	TransportType string `json:"transportType"`
	URL           string `json:"url"`
}

// McpCertifiedServerServerSpec defines the server configuration for McpCertifiedServer.
// It can be either a container image or a remote service.
type McpCertifiedServerServerSpec struct {
	// Type discriminator to determine the server type
	// +kubebuilder:validation:Required
	Type ServerType `json:"type"`

	// Container configuration (required when type is "container")
	// +optional
	Container *ContainerServerSpec `json:"container,omitempty"`

	// Remote configuration (required when type is "remote")
	// +optional
	Remote *RemoteServerSpec `json:"remote,omitempty"`
}

// McpCertifiedServerSpec defines the desired state of McpCertifiedServer.
type McpCertifiedServerSpec struct {
	McpServer McpCertifiedServerServerSpec `json:"mcpServer"`
}

// McpCertifiedServerStatus defines the observed state of McpCertifiedServer.
type McpCertifiedServerStatus struct {
	CertificationStatus CertificationStatus `json:"certificationStatus,omitempty"`
	ImageBuildStatus    ImageBuildStatus    `json:"imageBuildStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// McpCertifiedServer is the Schema for the mcpcertifiedservers API.
type McpCertifiedServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   McpCertifiedServerSpec   `json:"spec,omitempty"`
	Status McpCertifiedServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// McpCertifiedServerList contains a list of McpCertifiedServer.
type McpCertifiedServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []McpCertifiedServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&McpCertifiedServer{}, &McpCertifiedServerList{})
}
