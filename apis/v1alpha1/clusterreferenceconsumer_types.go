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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=crc,scope=Cluster
// +kubebuilder:metadata:annotations=api-approved.kubernetes.io=unapproved
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:storageversion

// ClusterReferenceConsumer identifies a consumer and its types of references.
// For example, a consumer may support references from Gateways to Secrets for
// tls-serving and Gateways to ConfigMaps for tls-client-validation.
type ClusterReferenceConsumer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Subject refers to the subject that is a consumer of the referenced
	// pattern(s).
	Subject Subject `json:"subject"`

	// ClassNames is an optional list of applicable classes for this Consumer if
	// the "From" API is partitioned by class
	ClassNames []string `json:"classNames,omitempty"`

	// References describe all of the resources a consumer may refer to
	References []ConsumerReference `json:"references"`
}

// ConsumerReference describes from which originating GroupResource to which
// target GroupResource a reference is for and for what purpose
type ConsumerReference struct {
	// From refers to the group and resource that these references originate from.
	From GroupResource `json:"from"`

	// To refers to the group and resource that these references target.
	To GroupResource `json:"to"`

	// For refers to the purpose of this reference. (Cluster)ReferenceGrants
	// matching the From, To, and For of this resource will be authorized for
	// the Subject of this resource.
	//
	// This value must be a valid DNS label as defined per RFC-1035.
	For string `json:"for"`
}

// +kubebuilder:object:root=true

// ClusterReferenceConsumerList contains a list of ClusterReferenceConsumer
type ClusterReferenceConsumerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterReferenceConsumer `json:"items"`
}

// Subject is a copy of RBAC Subject that excludes APIGroup.
type Subject struct {
	// Kind of object being referenced. Values defined by this API group are
	// "User", "Group", and "ServiceAccount". If the Authorizer does not
	// recognized the kind value, the Authorizer should report an error.
	Kind string `json:"kind"`
	// Name of the object being referenced.
	Name string `json:"name"`
	// Namespace of the referenced object.  If the object kind is non-namespace,
	// such as "User" or "Group", and this value is not empty the Authorizer
	// should report an error. +optional
	Namespace string `json:"namespace,omitempty"`
}
