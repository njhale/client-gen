/*
Copyright 2022 The KCP Authors.

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

package informergen

import (
	"io"
	"text/template"

	"github.com/kcp-dev/code-generator/pkg/util"
)

var (
	templateFuncs = template.FuncMap{
		"upper": util.UpperFirst,
		"lower": util.LowerFirst,
	}
)

type API struct {
	Group   string
	Version string
	Kind    string
	Package string
}

type genericInformer struct {
	PackageName string
	APIs        []API
}

func (g *genericInformer) WriteContent(w io.Writer) error {
	templ, err := template.New("generic").Funcs(templateFuncs).Parse(genericInformerTemplate)
	if err != nil {
		return err
	}
	return templ.Execute(w, g)
}

func NewGenericInformer(packageName string, apis []API) *genericInformer {
	return &genericInformer{APIs: apis, PackageName: packageName}
}

type groupInterface struct {
	PackageName     string
	InformerPackage string
	Group           string
	Versions        []string
}

func (g *groupInterface) WriteContent(w io.Writer) error {
	templ, err := template.New("groupInterface").Funcs(templateFuncs).Parse(groupInterfaceTemplate)
	if err != nil {
		return err
	}
	return templ.Execute(w, g)
}

func NewGroupInterface(packageName, informerPackage, group string, versions []string) *groupInterface {
	return &groupInterface{PackageName: packageName, InformerPackage: informerPackage, Group: group, Versions: versions}
}

type versionInterface struct {
	PackageName     string
	InformerPackage string
	Group           string
	Version         string
	APIs            []API
}

func (v *versionInterface) WriteContent(w io.Writer) error {
	templ, err := template.New("versionInterface").Funcs(templateFuncs).Parse(versionInterfaceTemplate)
	if err != nil {
		return err
	}
	return templ.Execute(w, v)
}

func NewVersionInterface(packageName, informerPackage, group, version string, apis []API) *versionInterface {
	return &versionInterface{PackageName: packageName, InformerPackage: informerPackage, Group: group, Version: version, APIs: apis}
}

type informer struct {
	PackageName     string
	InformerPackage string
	ListerPackage   string
	API             API
}

func (i *informer) WriteContent(w io.Writer) error {
	templ, err := template.New("informer").Funcs(templateFuncs).Parse(informerTemplate)
	if err != nil {
		return err
	}
	return templ.Execute(w, i)
}

func NewInformer(packageName, informerPackage, listerPackage string, api API) *informer {
	return &informer{PackageName: packageName, InformerPackage: informerPackage, ListerPackage: listerPackage, API: api}
}
