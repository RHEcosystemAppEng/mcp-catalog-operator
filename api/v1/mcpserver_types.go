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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RegistryRef struct {
	Name      string  `json:"name"`
	Namespace *string `json:"namespace,omitempty"`
}

type BlueprintRef struct {
	Name      string  `json:"name"`
	Namespace *string `json:"namespace,omitempty"`
}

type McpServerConfig struct {
	Guardian *bool `json:"guardian,omitempty"`

	// `blueprint` mode
	BlueprintRef *BlueprintRef `json:"blueprint-ref,omitempty"`
	Command      string        `json:"command,omitempty"`
	Args         []string      `json:"args,omitempty"`

	// `container` mode
	Proxy       *bool  `json:"proxy,omitempty"`
	ServerImage string `json:"server-image,omitempty"`

	// `remote` mode
	ServerURI string `json:"server-uri,omitempty"`
}

type McpServerSpec struct {
	Replicas    *int32                 `json:"replicas,omitempty"`
	EnvFrom     []corev1.EnvFromSource `json:"envFrom,omitempty"`
	RegistryRef RegistryRef            `json:"registry-ref,omitempty"`
	ServerMode  string                 `json:"server-mode"` // must be "blueprint", "container" or "remote"
	McpServer   McpServerConfig        `json:"mcp-server"`
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
