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

// McpPromotionJobSpec defines the desired state of McpPromotionJob.
type McpPromotionJobSpec struct {
	CatalogRef CatalogRef `json:"catalogRef"`

	Servers []string `json:"servers"`
}

// PromotionStatus is the state of a server promotion
// +kubebuilder:validation:Enum=planned;succeeded;failed
// +kubebuilder:object:generate=true
// +kubebuilder:validation:Optional
// +kubebuilder:default:=""
type PromotionStatus string

const (
	PromotionStatusPlanned   PromotionStatus = "planned"
	PromotionStatusSucceeded PromotionStatus = "succeeded"
	PromotionStatusFailed    PromotionStatus = "failed"
)

// ServerPromotionStatusDefinition defines the promotion status for a single server
// +kubebuilder:object:generate=true
// +kubebuilder:validation:Optional
// +kubebuilder:default:={}
type ServerPromotionStatusDefinition struct {
	Name             string          `json:"name"`
	PromotionStatus  PromotionStatus `json:"promotionStatus"`
	OriginalImage    string          `json:"originalImage"`
	DestinationImage string          `json:"destinationImage"`
}

// McpPromotionJobStatus defines the observed state of McpPromotionJob.
type McpPromotionJobStatus struct {
	ServerPromotions []ServerPromotionStatusDefinition `json:"serverPromotions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// McpPromotionJob is the Schema for the mcppromotionjob API.
type McpPromotionJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   McpPromotionJobSpec   `json:"spec,omitempty"`
	Status McpPromotionJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// McpPromotionJobList contains a list of McpPromotionJob.
type McpPromotionJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []McpPromotionJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&McpPromotionJob{}, &McpPromotionJobList{})
}
