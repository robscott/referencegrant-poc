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
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=crg
// +kubebuilder:metadata:annotations=api-approved.kubernetes.io=unapproved
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:storageversion

// ClusterReferenceGrant identifies a common form of referencing pattern. This
// can then be used with ReferenceGrants to selectively allow references.
type ClusterReferenceGrant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// From refers to the group and resource that these references originate
	// from.
	From GroupResource `json:"from"`

	// Versions describes how references and class partitions are defined for
	// the "From" API. Each Version string must be unique.
	Versions []VersionedReferencePaths `json:"versions"`
}

type VersionedReferencePaths struct {
	Version string `json:"version"`

	// ClassPath optionally refers to a field within an API that partitions it
	// on className.
	// It is an optional way to partition API access to consumers by their
	// applicable classNames.
	ClassPath string `json:"classPath,omitempty"`

	References []ReferencePath `json:"references"`
}

type ReferencePath struct {
	// Path in the "From" API where referenced names come from.
	Path string `json:"path"`

	// GroupResource for the target names from the Path
	To GroupResource `json:"to"`

	// For refers to the purpose of this reference. Subjects of
	// ClusterReferenceConsumers will be authorized to follow references
	// matching the From, To, and For of this resource.
	//
	// This value must be a valid DNS label as defined per RFC-1035.
	For string `json:"for"`
}

// +kubebuilder:object:root=true

// ClusterReferenceGrantList contains a list of ClusterReferenceGrant
type ClusterReferenceGrantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterReferenceGrant `json:"items"`
}
