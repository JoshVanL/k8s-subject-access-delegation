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
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type DestinationSubject struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type EventTrigger struct {
	Kind     string `json:"kind"`
	Value    string `json:"value"`
	Replicas int    `json:"replicas"`
}

type SubjectAccessDelegationSpec struct {
	DeletionTime string `json:"deletionTime"`
	Repeat       int    `json:"repeat"`

	OriginSubject       OriginSubject        `json:"originSubject"`
	DestinationSubjects []DestinationSubject `json:"destinationSubjects"`
	EventTriggers       []EventTrigger       `json:"triggers"`
}

type SubjectAccessDelegationStatus struct {
	Processed bool `json:"processed"`
	Triggerd  bool
	Iteration int

	RoleBindings        []string
	ClusterRoleBindings []string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=subjectaccessdelegationlist

type SubjectAccessDelegationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []SubjectAccessDelegation `json:"items"`
}
