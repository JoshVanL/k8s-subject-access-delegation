package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=subjectaccessdelegation

type SubjectAccessDelegation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubjectAccessDelegationSpec   `json:"spec"`
	Status SubjectAccessDelegationStatus `json:"status"`
}

type SubjectAccessDelegationSpec struct {
	User      string `json:"user"`
	Duration  string `json:"duration"`
	SomeStuff string `json:"somestuff"`
}

type SubjectAccessDelegationStatus struct {
	Processed bool `json:"processed"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=subjectaccessdelegationlist

type SubjectAccessDelegationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []SubjectAccessDelegation `json:"items"`
}
