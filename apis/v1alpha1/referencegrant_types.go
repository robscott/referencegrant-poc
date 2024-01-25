/*
Copyright 2023 The Kubernetes Authors.

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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=rg
// +kubebuilder:metadata:annotations=api-approved.kubernetes.io=unapproved
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:storageversion

// ReferenceGrant identifies namespaces of resources that are trusted to
// reference the specified names of resources in the same namespace as the
// grant.
type ReferenceGrant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// From describes the trusted namespaces and kinds that can reference the
	// resources described in the Pattern and optionally the "to" list.
	From GroupResourceNamespace `json:"from"`

	// To describes the names of resources that may be referenced from the
	// namespaces described in "From" following the linked pattern. When
	// unspecified or empty, references to all resources matching the pattern
	// are allowed.
	To ReferenceGrantTo `json:"to"`

	For For `json:"for"`
}

// +kubebuilder:object:root=true

// ReferenceGrantList contains a list of ReferenceGrant
type ReferenceGrantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReferenceGrant `json:"items"`
}

// ReferenceGrantTo describes what Names are allowed as targets of the
// references.
type ReferenceGrantTo struct {
	// Group is the group of the referents.
	Group string `json:"group"`

	// Resource is the resource of the referents.
	Resource string `json:"resource"`

	// Names are the names of the referents. When unspecified or empty, no
	// access is granted.
	//
	// +kubebuilder:validation:MaxItems=16
	Names []string `json:"names"`
}
