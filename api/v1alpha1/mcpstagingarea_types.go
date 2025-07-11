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

// McpStagingAreaSpec defines the desired state of McpStagingArea.
type McpStagingAreaSpec struct {
}

// McpStagingAreaStatus defines the observed state of McpStagingArea.
type McpStagingAreaStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// McpStagingArea is the Schema for the mcpstagingarea API.
type McpStagingArea struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   McpStagingAreaSpec   `json:"spec,omitempty"`
	Status McpStagingAreaStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// McpStagingAreaList contains a list of McpStagingArea.
type McpStagingAreaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []McpStagingArea `json:"items"`
}

func init() {
	SchemeBuilder.Register(&McpStagingArea{}, &McpStagingAreaList{})
}
