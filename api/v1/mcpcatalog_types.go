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

// McpCatalogSpec defines the desired state of McpCatalog.
type McpCatalogSpec struct {
	Description string `json:"description"`
}

// McpCatalogStatus defines the observed state of McpCatalog.
type McpCatalogStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// McpCatalog is the Schema for the mcpcatalogs API.
type McpCatalog struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   McpCatalogSpec   `json:"spec,omitempty"`
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
