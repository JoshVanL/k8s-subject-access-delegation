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

// Deprecated: register deep-copy functions.
func init() {
	SchemeBuilder.Register(RegisterDeepCopies)
}

// Deprecated: RegisterDeepCopies adds deep-copy functions to the given scheme. Public
// to allow building arbitrary schemes.
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

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, creating a new SubjectAccessDelegation.
func (x *SubjectAccessDelegation) DeepCopy() *SubjectAccessDelegation {
	if x == nil {
		return nil
	}
	out := new(SubjectAccessDelegation)
	x.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (x *SubjectAccessDelegation) DeepCopyObject() runtime.Object {
	if c := x.DeepCopy(); c != nil {
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

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, creating a new SubjectAccessDelegationList.
func (x *SubjectAccessDelegationList) DeepCopy() *SubjectAccessDelegationList {
	if x == nil {
		return nil
	}
	out := new(SubjectAccessDelegationList)
	x.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (x *SubjectAccessDelegationList) DeepCopyObject() runtime.Object {
	if c := x.DeepCopy(); c != nil {
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

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, creating a new SubjectAccessDelegationSpec.
func (x *SubjectAccessDelegationSpec) DeepCopy() *SubjectAccessDelegationSpec {
	if x == nil {
		return nil
	}
	out := new(SubjectAccessDelegationSpec)
	x.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubjectAccessDelegationStatus) DeepCopyInto(out *SubjectAccessDelegationStatus) {
	*out = *in
	return
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, creating a new SubjectAccessDelegationStatus.
func (x *SubjectAccessDelegationStatus) DeepCopy() *SubjectAccessDelegationStatus {
	if x == nil {
		return nil
	}
	out := new(SubjectAccessDelegationStatus)
	x.DeepCopyInto(out)
	return out
}
