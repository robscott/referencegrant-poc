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

	// PatternName refers to the name of the ClusterReferencePattern this allows.
	PatternName string `json:"patternName"`

	// From describes the trusted namespaces and kinds that can reference the
	// resources described in the Pattern and optionally the "to" list.
	//
	// Support: Core
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	From []ReferenceGrantFrom `json:"from"`

	// To describes the names of resources that may be referenced from the
	// namespaces described in "From" following the linked pattern. When
	// unspecified or empty, references to all resources matching the pattern
	// are allowed.
	//
	// +kubebuilder:validation:MaxItems=16
	To []ReferenceGrantTo `json:"to"`
}

// +kubebuilder:object:root=true

// ReferenceGrantList contains a list of ReferenceGrant
type ReferenceGrantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReferenceGrant `json:"items"`
}

// ReferenceGrantFrom describes trusted namespaces.
type ReferenceGrantFrom struct {
	// Namespace is the namespace of the referent.
	//
	// Support: Core
	Namespace string `json:"namespace"`
}

// ReferenceGrantTo describes what Names are allowed as targets of the
// references.
type ReferenceGrantTo struct {
	// Group is the group of the referent.
	Group string `json:"group"`

	// Resource is the resource of the referent.
	Resource string `json:"resource"`

	// Name is the name of the referent. When unspecified, this policy
	// refers to all resources of the specified Group and Kind in the local
	// namespace.
	//
	// +optional
	Name string `json:"name,omitempty"`
}
