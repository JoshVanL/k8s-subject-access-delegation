package authz

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SubjectAccessDelegation struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   SubjectAccessDelegationSpec
	Status SubjectAccessDelegationStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SubjectAccessDelegationList struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Items []SubjectAccessDelegation
}

type SubjectAccessDelegationSpec struct {
	User      string
	Duration  string
	SomeStuff string
}

type SubjectAccessDelegationStatus struct {
	Processed bool
}
