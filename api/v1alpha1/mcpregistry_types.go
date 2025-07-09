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

type RegistryRef struct {
	Name      string  `json:"name"`
	Namespace *string `json:"namespace,omitempty"`
}

// McpRegistrySpec defines the desired state of McpRegistry.
type McpRegistrySpec struct {
	Description string       `json:"description"`
	Catalogs    []CatalogRef `json:"catalogs"`
}

// McpRegistryStatus defines the observed state of McpRegistry.
type McpRegistryStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// McpRegistry is the Schema for the mcpregistries API.
type McpRegistry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   McpRegistrySpec   `json:"spec,omitempty"`
	Status McpRegistryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// McpRegistryList contains a list of McpRegistry.
type McpRegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []McpRegistry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&McpRegistry{}, &McpRegistryList{})
}
