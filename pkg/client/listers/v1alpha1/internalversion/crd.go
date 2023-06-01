/*
Copyright Saurov.

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
// Code generated by lister-gen. DO NOT EDIT.

package internalversion

import (
	v1alpha1 "github.com/souravbiswassanto/crd/pkg/apis/makecrd.com/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// CrdLister helps list Crds.
// All objects returned here must be treated as read-only.
type CrdLister interface {
	// List lists all Crds in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Crd, err error)
	// Crds returns an object that can list and get Crds.
	Crds(namespace string) CrdNamespaceLister
	CrdListerExpansion
}

// crdLister implements the CrdLister interface.
type crdLister struct {
	indexer cache.Indexer
}

// NewCrdLister returns a new CrdLister.
func NewCrdLister(indexer cache.Indexer) CrdLister {
	return &crdLister{indexer: indexer}
}

// List lists all Crds in the indexer.
func (s *crdLister) List(selector labels.Selector) (ret []*v1alpha1.Crd, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Crd))
	})
	return ret, err
}

// Crds returns an object that can list and get Crds.
func (s *crdLister) Crds(namespace string) CrdNamespaceLister {
	return crdNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// CrdNamespaceLister helps list and get Crds.
// All objects returned here must be treated as read-only.
type CrdNamespaceLister interface {
	// List lists all Crds in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Crd, err error)
	// Get retrieves the Crd from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.Crd, error)
	CrdNamespaceListerExpansion
}

// crdNamespaceLister implements the CrdNamespaceLister
// interface.
type crdNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Crds in the indexer for a given namespace.
func (s crdNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Crd, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Crd))
	})
	return ret, err
}

// Get retrieves the Crd from the indexer for a given namespace and name.
func (s crdNamespaceLister) Get(name string) (*v1alpha1.Crd, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("crd"), name)
	}
	return obj.(*v1alpha1.Crd), nil
}