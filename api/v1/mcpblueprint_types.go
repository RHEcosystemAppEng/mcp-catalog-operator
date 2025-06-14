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

type CatalogRef struct {
	Name      string  `json:"name"`
	Namespace *string `json:"namespace,omitempty"`
}

type McpBlueprintServerSpec struct {
	Proxy   *bool    `json:"proxy,omitempty"`
	Image   string   `json:"image"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	EnvVars []string `json:"env-vars"`
}

// McpBlueprintSpec defines the desired state of McpBlueprint.
type McpBlueprintSpec struct {
	CatalogRef   CatalogRef             `json:"catalog-ref,omitempty"`
	Description  string                 `json:"description"`
	Provider     string                 `json:"provider"`
	License      string                 `json:"license"`
	Competencies []string               `json:"competencies"`
	McpServer    McpBlueprintServerSpec `json:"mcp-server"`
}

// McpBlueprintStatus defines the observed state of McpBlueprint.
type McpBlueprintStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// McpBlueprint is the Schema for the mcpblueprints API.
type McpBlueprint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   McpBlueprintSpec   `json:"spec,omitempty"`
	Status McpBlueprintStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// McpBlueprintList contains a list of McpBlueprint.
type McpBlueprintList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []McpBlueprint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&McpBlueprint{}, &McpBlueprintList{})
}
