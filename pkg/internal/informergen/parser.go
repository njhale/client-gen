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
	"fmt"
	"io"
	"text/template"

	"github.com/kcp-dev/code-generator/pkg/util"
)

var (
	templateFuncs = template.FuncMap{
		"upper": util.UpperFirst,
		"lower": util.LowerFirst,
	}

	templates = map[string]string{
		Factory{}.Name():          factoryTemplate,
		GenericInformer{}.Name():  genericInformerTemplate,
		GroupInterface{}.Name():   groupInterfaceTemplate,
		VersionInterface{}.Name(): versionInterfaceTemplate,
		Informer{}.Name():         informerTemplate,
	}
)

type Parseable interface {
	Name() string
}

func WriteContent(w io.Writer, parseable Parseable) error {
	if w == nil {
		return fmt.Errorf("nil writer")
	}
	if parseable == nil {
		return fmt.Errorf("nil parseable")
	}

	name := parseable.Name()
	tmpl, ok := templates[parseable.Name()]
	if !ok {
		return fmt.Errorf("unknown parseable: %s", name)
	}

	parsed, err := template.New(name).Funcs(templateFuncs).Parse(tmpl)
	if err != nil {
		return err
	}

	return parsed.Execute(w, parseable)
}

type Factory struct {
	// TODO
	PackageName               string
	InformerPackage           string
	VersionedClientsetPackage string
}

func (Factory) Name() string {
	return "factory"
}

type API struct {
	Group   string
	Version string
	Kind    string
	Package string
}

type GenericInformer struct {
	PackageName string
	APIs        []API
}

func (GenericInformer) Name() string {
	return "genericInformer"
}

type GroupInterface struct {
	InformerPackage string
	Group           string
	Versions        []string
}

func (GroupInterface) Name() string {
	return "groupInterface"
}

type VersionInterface struct {
	InformerPackage string
	Group           string
	Version         string
	APIs            []API
}

func (VersionInterface) Name() string {
	return "versionInterface"
}

type Informer struct {
	InformerPackage string
	ListerPackage   string
	API
}

func (Informer) Name() string {
	return "informer"
}
