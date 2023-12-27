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
// +kubebuilder:resource:shortName=crp
// +kubebuilder:metadata:annotations=api-approved.kubernetes.io=unapproved
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:storageversion

// ClusterReferencePattern identifies a common form of referencing pattern. This
// can then be used with ReferenceGrants to selectively allow references.
type ClusterReferencePattern struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Group is the group of the referent.
	Group string `json:"group"`

	// Resource is the resource of the referent.
	Resource string `json:"resource"`

	// Paths are the path(s) which this reference may come from for each API
	// version.
	// +kubebuilder:validation:MinItems=1
	Paths []VersionedPath `json:"paths"`
}

// +kubebuilder:object:root=true

// ClusterReferencePatternList contains a list of ClusterReferencePattern
type ClusterReferencePatternList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterReferencePattern `json:"items"`
}

type VersionedPath struct {
	// Version is the API version of this resource this path applies to. When
	// unspecified, this path is used for all API versions.
	// +optional
	Version string `json:"apiVersion,omitempty"`

	// Path is the CEL path within the resource where the reference is defined.
	Path string `json:"path"`
}
