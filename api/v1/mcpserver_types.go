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

type RegistryRef struct {
	Name      string  `json:"name"`
	Namespace *string `json:"namespace,omitempty"`
}

type McpServerServerSpec struct {
	Proxy   *bool    `json:"proxy,omitempty"`
	Image   string   `json:"image"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	EnvVars []string `json:"envVars"`
}

// McpServerSpec defines the desired state of McpServer.
type McpServerSpec struct {
	RegistryRef  RegistryRef         `json:"registryRef,omitempty"`
	Description  string              `json:"description"`
	Provider     string              `json:"provider"`
	License      string              `json:"license"`
	Competencies []string            `json:"competencies"`
	McpServer    McpServerServerSpec `json:"mcpServer"`
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
