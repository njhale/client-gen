//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright The KCP Authors.

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

// Code generated by kcp code-generator. DO NOT EDIT.

package v1

import (
	apimachinerycache "github.com/kcp-dev/apimachinery/pkg/cache"
	examplev1 "github.com/kcp-dev/code-generator/examples/pkg/apis/example/v1"
	"github.com/kcp-dev/logicalcluster"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// TestTypeLister helps list examplev1.TestType.
// All objects returned here must be treated as read-only.
type TestTypeClusterLister interface {
	// List lists all examplev1.TestType in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*examplev1.TestType, err error)

	// Cluster returns an object that can list and get examplev1.TestType from the given logical cluster.
	Cluster(cluster logicalcluster.Name) TestTypeLister
}

// testTypeClusterLister implements the TestTypeClusterLister interface.
type testTypeClusterLister struct {
	indexer cache.Indexer
}

// NewTestTypeClusterLister returns a new TestTypeClusterLister.
func NewTestTypeClusterLister(indexer cache.Indexer) TestTypeClusterLister {
	return &testTypeClusterLister{indexer: indexer}
}

// List lists all examplev1.TestType in the indexer.
func (s *testTypeClusterLister) List(selector labels.Selector) (ret []*examplev1.TestType, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*examplev1.TestType))
	})
	return ret, err
}

// Cluster returns an object that can list and get examplev1.TestType.
func (s *testTypeClusterLister) Cluster(cluster logicalcluster.Name) TestTypeLister {
	return &testTypeLister{indexer: s.indexer, cluster: cluster}
}

// TestTypeLister helps list examplev1.TestType.
// All objects returned here must be treated as read-only.
type TestTypeLister interface {
	// List lists all examplev1.TestType in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*examplev1.TestType, err error)
	// TestTypes returns an object that can list and get examplev1.TestType.
	TestTypes(namespace string) TestTypeNamespaceLister
}

// testTypeLister implements the TestTypeLister interface.
type testTypeLister struct {
	indexer cache.Indexer
	cluster logicalcluster.Name
}

// List lists all examplev1.TestType in the indexer.
func (s *testTypeLister) List(selector labels.Selector) (ret []*examplev1.TestType, err error) {
	selectAll := selector == nil || selector.Empty()

	key := apimachinerycache.ToClusterAwareKey(s.cluster.String(), "", "")
	list, err := s.indexer.ByIndex(apimachinerycache.ClusterIndexName, key)
	if err != nil {
		return nil, err
	}

	for i := range list {
		obj := list[i].(*examplev1.TestType)
		if selectAll {
			ret = append(ret, obj)
		} else {
			if selector.Matches(labels.Set(obj.GetLabels())) {
				ret = append(ret, obj)
			}
		}
	}

	return ret, err
}

// TestTypes returns an object that can list and get examplev1.TestType.
func (s *testTypeLister) TestTypes(namespace string) TestTypeNamespaceLister {
	return testTypeNamespaceLister{indexer: s.indexer, cluster: s.cluster, namespace: namespace}
}

// TestTypeNamespaceLister helps list and get examplev1.TestType.
// All objects returned here must be treated as read-only.
type TestTypeNamespaceLister interface {
	// List lists all examplev1.TestType in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*examplev1.TestType, err error)
	// Get retrieves the examplev1.TestType from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*examplev1.TestType, error)
}

// testTypeNamespaceLister implements the TestTypeNamespaceLister interface.
type testTypeNamespaceLister struct {
	indexer   cache.Indexer
	cluster   logicalcluster.Name
	namespace string
}

// List lists all examplev1.TestType in the indexer for a given namespace.
func (s testTypeNamespaceLister) List(selector labels.Selector) (ret []*examplev1.TestType, err error) {
	selectAll := selector == nil || selector.Empty()

	key := apimachinerycache.ToClusterAwareKey(s.cluster.String(), s.namespace, "")
	list, err := s.indexer.ByIndex(apimachinerycache.ClusterAndNamespaceIndexName, key)
	if err != nil {
		return nil, err
	}

	for i := range list {
		obj := list[i].(*examplev1.TestType)
		if selectAll {
			ret = append(ret, obj)
		} else {
			if selector.Matches(labels.Set(obj.GetLabels())) {
				ret = append(ret, obj)
			}
		}
	}
	return ret, err
}

// Get retrieves the examplev1.TestType from the indexer for a given namespace and name.
func (s testTypeNamespaceLister) Get(name string) (*examplev1.TestType, error) {
	key := apimachinerycache.ToClusterAwareKey(s.cluster.String(), s.namespace, name)
	obj, exists, err := s.indexer.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(examplev1.Resource("testType"), name)
	}
	return obj.(*examplev1.TestType), nil
}
