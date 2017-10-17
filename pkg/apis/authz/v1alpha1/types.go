package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient=true
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=SubjectAccessDelegation

type SubjectAccessDelegation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubjectAccessDelegationSpec   `json:"spec,omitempty"`
	Status SubjectAccessDelegationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SubjectAccessDelegationList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Items []SubjectAccessDelegation `json:"items"`
}

type SubjectAccessDelegationSpec struct {
	User      string `json:"user"`
	Duration  string `json:"duration"`
	SomeStuff string `json:"somestuff"`
}

type SubjectAccessDelegationStatus struct {
	Processed bool `json:"processed"`
}
