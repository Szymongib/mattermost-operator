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

package v1alpha1

import (
	"time"

	v1alpha1 "github.com/mattermost/mattermost-operator/pkg/apis/mattermost/v1alpha1"
	scheme "github.com/mattermost/mattermost-operator/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ClusterInstallationsGetter has a method to return a ClusterInstallationInterface.
// A group's client should implement this interface.
type ClusterInstallationsGetter interface {
	ClusterInstallations(namespace string) ClusterInstallationInterface
}

// ClusterInstallationInterface has methods to work with ClusterInstallation resources.
type ClusterInstallationInterface interface {
	Create(*v1alpha1.ClusterInstallation) (*v1alpha1.ClusterInstallation, error)
	Update(*v1alpha1.ClusterInstallation) (*v1alpha1.ClusterInstallation, error)
	UpdateStatus(*v1alpha1.ClusterInstallation) (*v1alpha1.ClusterInstallation, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.ClusterInstallation, error)
	List(opts v1.ListOptions) (*v1alpha1.ClusterInstallationList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ClusterInstallation, err error)
	ClusterInstallationExpansion
}

// clusterInstallations implements ClusterInstallationInterface
type clusterInstallations struct {
	client rest.Interface
	ns     string
}

// newClusterInstallations returns a ClusterInstallations
func newClusterInstallations(c *MattermostV1alpha1Client, namespace string) *clusterInstallations {
	return &clusterInstallations{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the clusterInstallation, and returns the corresponding clusterInstallation object, and an error if there is any.
func (c *clusterInstallations) Get(name string, options v1.GetOptions) (result *v1alpha1.ClusterInstallation, err error) {
	result = &v1alpha1.ClusterInstallation{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("clusterinstallations").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ClusterInstallations that match those selectors.
func (c *clusterInstallations) List(opts v1.ListOptions) (result *v1alpha1.ClusterInstallationList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ClusterInstallationList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("clusterinstallations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested clusterInstallations.
func (c *clusterInstallations) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("clusterinstallations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a clusterInstallation and creates it.  Returns the server's representation of the clusterInstallation, and an error, if there is any.
func (c *clusterInstallations) Create(clusterInstallation *v1alpha1.ClusterInstallation) (result *v1alpha1.ClusterInstallation, err error) {
	result = &v1alpha1.ClusterInstallation{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("clusterinstallations").
		Body(clusterInstallation).
		Do().
		Into(result)
	return
}

// Update takes the representation of a clusterInstallation and updates it. Returns the server's representation of the clusterInstallation, and an error, if there is any.
func (c *clusterInstallations) Update(clusterInstallation *v1alpha1.ClusterInstallation) (result *v1alpha1.ClusterInstallation, err error) {
	result = &v1alpha1.ClusterInstallation{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("clusterinstallations").
		Name(clusterInstallation.Name).
		Body(clusterInstallation).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *clusterInstallations) UpdateStatus(clusterInstallation *v1alpha1.ClusterInstallation) (result *v1alpha1.ClusterInstallation, err error) {
	result = &v1alpha1.ClusterInstallation{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("clusterinstallations").
		Name(clusterInstallation.Name).
		SubResource("status").
		Body(clusterInstallation).
		Do().
		Into(result)
	return
}

// Delete takes name of the clusterInstallation and deletes it. Returns an error if one occurs.
func (c *clusterInstallations) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("clusterinstallations").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *clusterInstallations) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("clusterinstallations").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched clusterInstallation.
func (c *clusterInstallations) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ClusterInstallation, err error) {
	result = &v1alpha1.ClusterInstallation{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("clusterinstallations").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}