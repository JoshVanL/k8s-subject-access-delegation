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

package v1alpha1

import (
	v1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	scheme "github.com/joshvanl/k8s-subject-access-delegation/pkg/client/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// SubjectAccessDelegationsGetter has a method to return a SubjectAccessDelegationInterface.
// A group's client should implement this interface.
type SubjectAccessDelegationsGetter interface {
	SubjectAccessDelegations(namespace string) SubjectAccessDelegationInterface
}

// SubjectAccessDelegationInterface has methods to work with SubjectAccessDelegation resources.
type SubjectAccessDelegationInterface interface {
	Create(*v1alpha1.SubjectAccessDelegation) (*v1alpha1.SubjectAccessDelegation, error)
	Update(*v1alpha1.SubjectAccessDelegation) (*v1alpha1.SubjectAccessDelegation, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.SubjectAccessDelegation, error)
	List(opts v1.ListOptions) (*v1alpha1.SubjectAccessDelegationList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.SubjectAccessDelegation, err error)
	SubjectAccessDelegationExpansion
}

// subjectAccessDelegations implements SubjectAccessDelegationInterface
type subjectAccessDelegations struct {
	client rest.Interface
	ns     string
}

// newSubjectAccessDelegations returns a SubjectAccessDelegations
func newSubjectAccessDelegations(c *AuthzV1alpha1Client, namespace string) *subjectAccessDelegations {
	return &subjectAccessDelegations{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the subjectAccessDelegation, and returns the corresponding subjectAccessDelegation object, and an error if there is any.
func (c *subjectAccessDelegations) Get(name string, options v1.GetOptions) (result *v1alpha1.SubjectAccessDelegation, err error) {
	result = &v1alpha1.SubjectAccessDelegation{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("subjectaccessdelegations").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SubjectAccessDelegations that match those selectors.
func (c *subjectAccessDelegations) List(opts v1.ListOptions) (result *v1alpha1.SubjectAccessDelegationList, err error) {
	result = &v1alpha1.SubjectAccessDelegationList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("subjectaccessdelegations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested subjectAccessDelegations.
func (c *subjectAccessDelegations) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("subjectaccessdelegations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a subjectAccessDelegation and creates it.  Returns the server's representation of the subjectAccessDelegation, and an error, if there is any.
func (c *subjectAccessDelegations) Create(subjectAccessDelegation *v1alpha1.SubjectAccessDelegation) (result *v1alpha1.SubjectAccessDelegation, err error) {
	result = &v1alpha1.SubjectAccessDelegation{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("subjectaccessdelegations").
		Body(subjectAccessDelegation).
		Do().
		Into(result)
	return
}

// Update takes the representation of a subjectAccessDelegation and updates it. Returns the server's representation of the subjectAccessDelegation, and an error, if there is any.
func (c *subjectAccessDelegations) Update(subjectAccessDelegation *v1alpha1.SubjectAccessDelegation) (result *v1alpha1.SubjectAccessDelegation, err error) {
	result = &v1alpha1.SubjectAccessDelegation{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("subjectaccessdelegations").
		Name(subjectAccessDelegation.Name).
		Body(subjectAccessDelegation).
		Do().
		Into(result)
	return
}

// Delete takes name of the subjectAccessDelegation and deletes it. Returns an error if one occurs.
func (c *subjectAccessDelegations) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("subjectaccessdelegations").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *subjectAccessDelegations) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("subjectaccessdelegations").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched subjectAccessDelegation.
func (c *subjectAccessDelegations) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.SubjectAccessDelegation, err error) {
	result = &v1alpha1.SubjectAccessDelegation{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("subjectaccessdelegations").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
