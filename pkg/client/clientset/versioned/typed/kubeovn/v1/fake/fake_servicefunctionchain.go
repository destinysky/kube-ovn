/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	kubeovnv1 "github.com/alauda/kube-ovn/pkg/apis/kubeovn/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeServiceFunctionChains implements ServiceFunctionChainInterface
type FakeServiceFunctionChains struct {
	Fake *FakeKubeovnV1
	ns   string
}

var servicefunctionchainsResource = schema.GroupVersionResource{Group: "kubeovn.io", Version: "v1", Resource: "servicefunctionchains"}

var servicefunctionchainsKind = schema.GroupVersionKind{Group: "kubeovn.io", Version: "v1", Kind: "ServiceFunctionChain"}

// Get takes name of the serviceFunctionChain, and returns the corresponding serviceFunctionChain object, and an error if there is any.
func (c *FakeServiceFunctionChains) Get(name string, options v1.GetOptions) (result *kubeovnv1.ServiceFunctionChain, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(servicefunctionchainsResource, c.ns, name), &kubeovnv1.ServiceFunctionChain{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubeovnv1.ServiceFunctionChain), err
}

// List takes label and field selectors, and returns the list of ServiceFunctionChains that match those selectors.
func (c *FakeServiceFunctionChains) List(opts v1.ListOptions) (result *kubeovnv1.ServiceFunctionChainList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(servicefunctionchainsResource, servicefunctionchainsKind, c.ns, opts), &kubeovnv1.ServiceFunctionChainList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &kubeovnv1.ServiceFunctionChainList{ListMeta: obj.(*kubeovnv1.ServiceFunctionChainList).ListMeta}
	for _, item := range obj.(*kubeovnv1.ServiceFunctionChainList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested serviceFunctionChains.
func (c *FakeServiceFunctionChains) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(servicefunctionchainsResource, c.ns, opts))

}

// Create takes the representation of a serviceFunctionChain and creates it.  Returns the server's representation of the serviceFunctionChain, and an error, if there is any.
func (c *FakeServiceFunctionChains) Create(serviceFunctionChain *kubeovnv1.ServiceFunctionChain) (result *kubeovnv1.ServiceFunctionChain, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(servicefunctionchainsResource, c.ns, serviceFunctionChain), &kubeovnv1.ServiceFunctionChain{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubeovnv1.ServiceFunctionChain), err
}

// Update takes the representation of a serviceFunctionChain and updates it. Returns the server's representation of the serviceFunctionChain, and an error, if there is any.
func (c *FakeServiceFunctionChains) Update(serviceFunctionChain *kubeovnv1.ServiceFunctionChain) (result *kubeovnv1.ServiceFunctionChain, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(servicefunctionchainsResource, c.ns, serviceFunctionChain), &kubeovnv1.ServiceFunctionChain{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubeovnv1.ServiceFunctionChain), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeServiceFunctionChains) UpdateStatus(serviceFunctionChain *kubeovnv1.ServiceFunctionChain) (*kubeovnv1.ServiceFunctionChain, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(servicefunctionchainsResource, "status", c.ns, serviceFunctionChain), &kubeovnv1.ServiceFunctionChain{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubeovnv1.ServiceFunctionChain), err
}

// Delete takes name of the serviceFunctionChain and deletes it. Returns an error if one occurs.
func (c *FakeServiceFunctionChains) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(servicefunctionchainsResource, c.ns, name), &kubeovnv1.ServiceFunctionChain{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeServiceFunctionChains) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(servicefunctionchainsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &kubeovnv1.ServiceFunctionChainList{})
	return err
}

// Patch applies the patch and returns the patched serviceFunctionChain.
func (c *FakeServiceFunctionChains) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kubeovnv1.ServiceFunctionChain, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(servicefunctionchainsResource, c.ns, name, pt, data, subresources...), &kubeovnv1.ServiceFunctionChain{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubeovnv1.ServiceFunctionChain), err
}
