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

import "TODO/example/v1"

type Interface interface {
	TestTypes() TestTypeInformer
	ClusterTestTypes() ClusterTestTypeInformer
}

type version struct {
	delegate v1.Interface
}

func New(delegate v1.Interface) Interface {
	return &version{delegate: delegate}
}

func (v *version) TestTypes() TestTypeInformer {
	return &testTypeInformer{
		delegate: v.delegate.TestTypes(),
	}
}
func (v *version) ClusterTestTypes() ClusterTestTypeInformer {
	return &clusterTestTypeInformer{
		delegate: v.delegate.ClusterTestTypes(),
	}
}
