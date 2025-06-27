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

type CatalogRef struct {
	Name      string  `json:"name"`
	Namespace *string `json:"namespace,omitempty"`
}

// McpCatalogSpec defines the desired state of McpCatalog.
// +kubebuilder:validation:XValidation:rule=has(self.description) && size(self.description) > 0 && has(self.imageRegistry) && size(self.imageRegistry) > 0, message="Both description and imageRegistry are required and cannot be empty."
type McpCatalogSpec struct {
	Description   string `json:"description"`
	ImageRegistry string `json:"imageRegistry"`
}

// McpCatalogStatus defines the observed state of McpCatalog.
type McpCatalogStatus struct {
	// Conditions represent the latest available observations of a McpCatalog's current state.
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// McpCatalog is the Schema for the mcpcatalogs API.
type McpCatalog struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   McpCatalogSpec   `json:"spec"`
	Status McpCatalogStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// McpCatalogList contains a list of McpCatalog.
type McpCatalogList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []McpCatalog `json:"items"`
}

func init() {
	SchemeBuilder.Register(&McpCatalog{}, &McpCatalogList{})
}
