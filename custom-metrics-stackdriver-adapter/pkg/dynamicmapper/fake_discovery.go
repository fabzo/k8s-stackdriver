/*
Copyright 2016 The Kubernetes Authors.

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

package dynamicmapper

import (
	"fmt"

	"github.com/emicklei/go-restful-swagger12"

	"github.com/go-openapi/spec"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/pkg/api/v1"
	kubeversion "k8s.io/client-go/pkg/version"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/testing"
)

// NB: this is a copy of k8s.io/client-go/discovery/fake.  The original returns `nil, nil`
// for some methods, which is generally confuses lots of code.

// FakeDiscovery is a fake implementation of DiscoveryClient.
type FakeDiscovery struct {
	*testing.Fake
}

// ServerResourcesForGroupVersion returns the supported resources for a group and version.
func (c *FakeDiscovery) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	action := testing.ActionImpl{
		Verb:     "get",
		Resource: schema.GroupVersionResource{Resource: "resource"},
	}
	c.Invokes(action, nil)
	for _, resourceList := range c.Resources {
		if resourceList.GroupVersion == groupVersion {
			return resourceList, nil
		}
	}
	return nil, fmt.Errorf("GroupVersion %q not found", groupVersion)
}

// ServerResources returns the supported resources for all groups and versions.
func (c *FakeDiscovery) ServerResources() ([]*metav1.APIResourceList, error) {
	action := testing.ActionImpl{
		Verb:     "get",
		Resource: schema.GroupVersionResource{Resource: "resource"},
	}
	c.Invokes(action, nil)
	return c.Resources, nil
}

// ServerPreferredResources returns the supported resources with the version preferred by the
// server.
func (c *FakeDiscovery) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	return nil, nil
}

// ServerPreferredNamespacedResources returns the supported namespaced resources with the
// version preferred by the server.
func (c *FakeDiscovery) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	return nil, nil
}

// ServerGroups returns the supported groups, with information like supported versions and the
// preferred version.
func (c *FakeDiscovery) ServerGroups() (*metav1.APIGroupList, error) {
	groups := map[string]*metav1.APIGroup{}
	groupVersions := map[metav1.GroupVersionForDiscovery]struct{}{}
	for _, resourceList := range c.Resources {
		groupVer, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			return nil, err
		}
		groupVerForDisc := metav1.GroupVersionForDiscovery{
			GroupVersion: resourceList.GroupVersion,
			Version:      groupVer.Version,
		}

		group, groupPresent := groups[groupVer.Group]
		if !groupPresent {
			group = &metav1.APIGroup{
				Name: groupVer.Group,
				// use the fist seen version as the preferred version
				PreferredVersion: groupVerForDisc,
			}
			groups[groupVer.Group] = group
		}

		// we'll dedup in the end by deleting the group-versions
		// from the global map one at a time
		group.Versions = append(group.Versions, groupVerForDisc)
		groupVersions[groupVerForDisc] = struct{}{}
	}

	groupList := make([]metav1.APIGroup, 0, len(groups))
	for _, group := range groups {
		newGroup := metav1.APIGroup{
			Name:             group.Name,
			PreferredVersion: group.PreferredVersion,
		}

		for _, groupVer := range group.Versions {
			if _, ok := groupVersions[groupVer]; ok {
				delete(groupVersions, groupVer)
				newGroup.Versions = append(newGroup.Versions, groupVer)
			}
		}

		groupList = append(groupList, newGroup)
	}

	return &metav1.APIGroupList{
		Groups: groupList,
	}, nil
}

// ServerVersion retrieves and parses the server's version (git version).
func (c *FakeDiscovery) ServerVersion() (*version.Info, error) {
	action := testing.ActionImpl{}
	action.Verb = "get"
	action.Resource = schema.GroupVersionResource{Resource: "version"}

	c.Invokes(action, nil)
	versionInfo := kubeversion.Get()
	return &versionInfo, nil
}

// SwaggerSchema retrieves and parses the swagger API schema the server supports.
func (c *FakeDiscovery) SwaggerSchema(version schema.GroupVersion) (*swagger.ApiDeclaration, error) {
	action := testing.ActionImpl{}
	action.Verb = "get"
	if version == v1.SchemeGroupVersion {
		action.Resource = schema.GroupVersionResource{Resource: "/swaggerapi/api/" + version.Version}
	} else {
		action.Resource = schema.GroupVersionResource{Resource: "/swaggerapi/apis/" + version.Group + "/" + version.Version}
	}

	c.Invokes(action, nil)
	return &swagger.ApiDeclaration{}, nil
}

// OpenAPISchema fetches the open api schema using a rest client and parses the proto.
func (c *FakeDiscovery) OpenAPISchema() (*spec.Swagger, error) { return &spec.Swagger{}, nil }

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeDiscovery) RESTClient() restclient.Interface {
	return nil
}
