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

package controller

const (
	ValidationMessageDescriptionRequired   = "description field is required and cannot be empty or null"
	ValidationMessageImageRegistryRequired = "imageRegistry field is required and cannot be empty or null"
	ValidationMessageCatalogSuccess        = "McpCatalog spec is valid"
	ValidationMessageCatalogNotFound       = "referenced catalog does not exist"
	ValidationMessageServerSuccess         = "McpServer spec is valid"
	ValidationMessageCrossNamespaces       = "referenced catalog is in a different namespace"
)

// Condition types
const (
	// ConditionTypeReady represents the Ready condition type
	ConditionTypeReady = "Ready"
)

// Condition reasons
const (
	ConditionReasonValidationSucceeded = "ValidationSucceeded"
	ConditionReasonValidationFailed    = "ValidationFailed"
	ConditionReasonCrossNamespaces     = "CrossNamespaces"
)
