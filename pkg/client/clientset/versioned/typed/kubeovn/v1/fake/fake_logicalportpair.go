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

// FakeLogicalPortPairs implements LogicalPortPairInterface
type FakeLogicalPortPairs struct {
	Fake *FakeKubeovnV1
	ns   string
}

var logicalportpairsResource = schema.GroupVersionResource{Group: "kubeovn.io", Version: "v1", Resource: "logicalportpairs"}

var logicalportpairsKind = schema.GroupVersionKind{Group: "kubeovn.io", Version: "v1", Kind: "LogicalPortPair"}

// Get takes name of the logicalPortPair, and returns the corresponding logicalPortPair object, and an error if there is any.
func (c *FakeLogicalPortPairs) Get(name string, options v1.GetOptions) (result *kubeovnv1.LogicalPortPair, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(logicalportpairsResource, c.ns, name), &kubeovnv1.LogicalPortPair{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubeovnv1.LogicalPortPair), err
}

// List takes label and field selectors, and returns the list of LogicalPortPairs that match those selectors.
func (c *FakeLogicalPortPairs) List(opts v1.ListOptions) (result *kubeovnv1.LogicalPortPairList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(logicalportpairsResource, logicalportpairsKind, c.ns, opts), &kubeovnv1.LogicalPortPairList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &kubeovnv1.LogicalPortPairList{ListMeta: obj.(*kubeovnv1.LogicalPortPairList).ListMeta}
	for _, item := range obj.(*kubeovnv1.LogicalPortPairList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested logicalPortPairs.
func (c *FakeLogicalPortPairs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(logicalportpairsResource, c.ns, opts))

}

// Create takes the representation of a logicalPortPair and creates it.  Returns the server's representation of the logicalPortPair, and an error, if there is any.
func (c *FakeLogicalPortPairs) Create(logicalPortPair *kubeovnv1.LogicalPortPair) (result *kubeovnv1.LogicalPortPair, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(logicalportpairsResource, c.ns, logicalPortPair), &kubeovnv1.LogicalPortPair{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubeovnv1.LogicalPortPair), err
}

// Update takes the representation of a logicalPortPair and updates it. Returns the server's representation of the logicalPortPair, and an error, if there is any.
func (c *FakeLogicalPortPairs) Update(logicalPortPair *kubeovnv1.LogicalPortPair) (result *kubeovnv1.LogicalPortPair, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(logicalportpairsResource, c.ns, logicalPortPair), &kubeovnv1.LogicalPortPair{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubeovnv1.LogicalPortPair), err
}

// Delete takes name of the logicalPortPair and deletes it. Returns an error if one occurs.
func (c *FakeLogicalPortPairs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(logicalportpairsResource, c.ns, name), &kubeovnv1.LogicalPortPair{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeLogicalPortPairs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(logicalportpairsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &kubeovnv1.LogicalPortPairList{})
	return err
}

// Patch applies the patch and returns the patched logicalPortPair.
func (c *FakeLogicalPortPairs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kubeovnv1.LogicalPortPair, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(logicalportpairsResource, c.ns, name, pt, data, subresources...), &kubeovnv1.LogicalPortPair{})

	if obj == nil {
		return nil, err
	}
	return obj.(*kubeovnv1.LogicalPortPair), err
}
