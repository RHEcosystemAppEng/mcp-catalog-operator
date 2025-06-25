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

// McpServerCertJobSpec defines the desired state of McpServerCertJob.
type McpServerCertJobSpec struct {
	// Reference to the McpServer to be certified
	ServerRef ServerRef `json:"serverRef,omitempty"`
}

// McpServerCertJobStatus defines the observed state of McpServerCertJob.
type McpServerCertJobStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// McpServerCertJob is the Schema for the mcpservercertjobs API.
type McpServerCertJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   McpServerCertJobSpec   `json:"spec,omitempty"`
	Status McpServerCertJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// McpServerCertJobList contains a list of McpServerCertJob.
type McpServerCertJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []McpServerCertJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&McpServerCertJob{}, &McpServerCertJobList{})
}
