// +build !ignore_autogenerated

/*
Copyright 2017 The Kubernetes Authors.

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

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package v1alpha1

import (
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
	reflect "reflect"
)

func init() {
	SchemeBuilder.Register(RegisterDeepCopies)
}

// RegisterDeepCopies adds deep-copy functions to the given scheme. Public
// to allow building arbitrary schemes.
//
// Deprecated: deepcopy registration will go away when static deepcopy is fully implemented.
func RegisterDeepCopies(scheme *runtime.Scheme) error {
	return scheme.AddGeneratedDeepCopyFuncs(
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*SubjectAccessDelegation).DeepCopyInto(out.(*SubjectAccessDelegation))
			return nil
		}, InType: reflect.TypeOf(&SubjectAccessDelegation{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*SubjectAccessDelegationList).DeepCopyInto(out.(*SubjectAccessDelegationList))
			return nil
		}, InType: reflect.TypeOf(&SubjectAccessDelegationList{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*SubjectAccessDelegationSpec).DeepCopyInto(out.(*SubjectAccessDelegationSpec))
			return nil
		}, InType: reflect.TypeOf(&SubjectAccessDelegationSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*SubjectAccessDelegationStatus).DeepCopyInto(out.(*SubjectAccessDelegationStatus))
			return nil
		}, InType: reflect.TypeOf(&SubjectAccessDelegationStatus{})},
	)
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubjectAccessDelegation) DeepCopyInto(out *SubjectAccessDelegation) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubjectAccessDelegation.
func (in *SubjectAccessDelegation) DeepCopy() *SubjectAccessDelegation {
	if in == nil {
		return nil
	}
	out := new(SubjectAccessDelegation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SubjectAccessDelegation) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubjectAccessDelegationList) DeepCopyInto(out *SubjectAccessDelegationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SubjectAccessDelegation, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubjectAccessDelegationList.
func (in *SubjectAccessDelegationList) DeepCopy() *SubjectAccessDelegationList {
	if in == nil {
		return nil
	}
	out := new(SubjectAccessDelegationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SubjectAccessDelegationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubjectAccessDelegationSpec) DeepCopyInto(out *SubjectAccessDelegationSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubjectAccessDelegationSpec.
func (in *SubjectAccessDelegationSpec) DeepCopy() *SubjectAccessDelegationSpec {
	if in == nil {
		return nil
	}
	out := new(SubjectAccessDelegationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubjectAccessDelegationStatus) DeepCopyInto(out *SubjectAccessDelegationStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubjectAccessDelegationStatus.
func (in *SubjectAccessDelegationStatus) DeepCopy() *SubjectAccessDelegationStatus {
	if in == nil {
		return nil
	}
	out := new(SubjectAccessDelegationStatus)
	in.DeepCopyInto(out)
	return out
}
