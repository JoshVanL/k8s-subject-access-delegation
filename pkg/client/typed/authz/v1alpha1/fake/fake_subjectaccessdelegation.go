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

package fake

import (
	v1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSubjectAccessDelegations implements SubjectAccessDelegationInterface
type FakeSubjectAccessDelegations struct {
	Fake *FakeAuthzV1alpha1
	ns   string
}

var subjectaccessdelegationsResource = schema.GroupVersionResource{Group: "authz", Version: "v1alpha1", Resource: "subjectaccessdelegations"}

var subjectaccessdelegationsKind = schema.GroupVersionKind{Group: "authz", Version: "v1alpha1", Kind: "SubjectAccessDelegation"}

func (c *FakeSubjectAccessDelegations) Create(subjectAccessDelegation *v1alpha1.SubjectAccessDelegation) (result *v1alpha1.SubjectAccessDelegation, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(subjectaccessdelegationsResource, c.ns, subjectAccessDelegation), &v1alpha1.SubjectAccessDelegation{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SubjectAccessDelegation), err
}

func (c *FakeSubjectAccessDelegations) Update(subjectAccessDelegation *v1alpha1.SubjectAccessDelegation) (result *v1alpha1.SubjectAccessDelegation, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(subjectaccessdelegationsResource, c.ns, subjectAccessDelegation), &v1alpha1.SubjectAccessDelegation{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SubjectAccessDelegation), err
}

func (c *FakeSubjectAccessDelegations) UpdateStatus(subjectAccessDelegation *v1alpha1.SubjectAccessDelegation) (*v1alpha1.SubjectAccessDelegation, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(subjectaccessdelegationsResource, "status", c.ns, subjectAccessDelegation), &v1alpha1.SubjectAccessDelegation{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SubjectAccessDelegation), err
}

func (c *FakeSubjectAccessDelegations) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(subjectaccessdelegationsResource, c.ns, name), &v1alpha1.SubjectAccessDelegation{})

	return err
}

func (c *FakeSubjectAccessDelegations) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(subjectaccessdelegationsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.SubjectAccessDelegationList{})
	return err
}

func (c *FakeSubjectAccessDelegations) Get(name string, options v1.GetOptions) (result *v1alpha1.SubjectAccessDelegation, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(subjectaccessdelegationsResource, c.ns, name), &v1alpha1.SubjectAccessDelegation{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SubjectAccessDelegation), err
}

func (c *FakeSubjectAccessDelegations) List(opts v1.ListOptions) (result *v1alpha1.SubjectAccessDelegationList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(subjectaccessdelegationsResource, subjectaccessdelegationsKind, c.ns, opts), &v1alpha1.SubjectAccessDelegationList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.SubjectAccessDelegationList{}
	for _, item := range obj.(*v1alpha1.SubjectAccessDelegationList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested subjectAccessDelegations.
func (c *FakeSubjectAccessDelegations) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(subjectaccessdelegationsResource, c.ns, opts))

}

// Patch applies the patch and returns the patched subjectAccessDelegation.
func (c *FakeSubjectAccessDelegations) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.SubjectAccessDelegation, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(subjectaccessdelegationsResource, c.ns, name, data, subresources...), &v1alpha1.SubjectAccessDelegation{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SubjectAccessDelegation), err
}
