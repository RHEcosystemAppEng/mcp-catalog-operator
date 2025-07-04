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

// +kubebuilder:validation:Enum=running;completed;failed
type ImportJobStatus string

const (
	ImportJobRunning   ImportJobStatus = "running"
	ImportJobCompleted ImportJobStatus = "completed"
	ImportJobFailed    ImportJobStatus = "failed"
)

// McpImportJobSpec defines the desired state of McpImportJob.
type McpImportJobSpec struct {
	// Reference to the MCP Registry to import the servers from
	RegistryURI string `json:"registryUri"`
	// TODO: Add support for authentication

	// Name filters to apply to the imported servers (optional, defaults to all)
	NameFilter *string `json:"nameFilter,omitempty"`
	// Maximum number of servers to import (optional, defaults to 10)
	MaxServers *int `json:"maxServers,omitempty"`
}

// McpImportJobStatus defines the observed state of McpImportJob.
type McpImportJobStatus struct {
	// Job status
	Status ImportJobStatus `json:"status"`
	// Name of the ConfigMap that contains the import details
	ConfigMapName *string `json:"configMapName,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// McpImportJob is the Schema for the mcpimportjobs API.
type McpImportJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   McpImportJobSpec   `json:"spec,omitempty"`
	Status McpImportJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// McpImportJobList contains a list of McpImportJob.
type McpImportJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []McpImportJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&McpImportJob{}, &McpImportJobList{})
}
