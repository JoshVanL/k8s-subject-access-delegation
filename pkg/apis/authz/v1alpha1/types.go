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

type OriginSubject struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type DestinationSubject struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type SubjectAccessDelegationSpec struct {
	Duration            int64                `json:"duration"`
	Repeat              int                  `json:"repeat"`
	Namespace           string               `json:"namespace"`
	OriginSubject       OriginSubject        `json:"originSubject"`
	DestinationSubjects []DestinationSubject `json:"destinationSubjects"`
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
