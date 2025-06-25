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

type McpCertifiedServerServerSpec struct {
	Proxy   *bool    `json:"proxy,omitempty"`
	Image   string   `json:"image"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	EnvVars []string `json:"envVars"`
}

// McpCertifiedServerSpec defines the desired state of McpCertifiedServer.
type McpCertifiedServerSpec struct {
	RegistryRef  RegistryRef                  `json:"registryRef,omitempty"`
	Description  string                       `json:"description"`
	Provider     string                       `json:"provider"`
	License      string                       `json:"license"`
	Competencies []string                     `json:"competencies"`
	McpServer    McpCertifiedServerServerSpec `json:"mcpServer"`
}

// McpCertifiedServerStatus defines the observed state of McpCertifiedServer.
type McpCertifiedServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
