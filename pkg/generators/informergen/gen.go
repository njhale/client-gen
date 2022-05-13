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
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"io"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
	"k8s.io/code-generator/cmd/client-gen/types"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"

	"github.com/kcp-dev/code-generator/pkg/flag"
	"github.com/kcp-dev/code-generator/pkg/generators/clientgen"
	"github.com/kcp-dev/code-generator/pkg/internal/informergen"
	"github.com/kcp-dev/code-generator/pkg/util"
)

const (
	// GeneratorName is the name of the generator.
	GeneratorName = "informer"
	// packageName for typed client wrappers.
	typedPackageName = "externalversions"
)

type Generator struct {
	// inputDir is the path where types are defined.
	inputDir string

	// baseInformersPackage is the base package of the informers to be made kcp-aware.
	baseInformersPackage string

	// output Dir where the wrappers are to be written.
	outputDir string

	// GroupVersions for whom the clients are to be generated.
	groupVersions []types.GroupVersions

	// headerText is the header text to be added to generated wrappers.
	// It is obtained from `--go-header-text` flag.
	headerText string
}

func (g Generator) RegisterMarker() (*markers.Registry, error) {
	reg := &markers.Registry{}
	if err := markers.RegisterAll(reg, RuleDefinition, NonNamespacedMarker, NoStatusMarker); err != nil {
		return nil, fmt.Errorf("error registering markers")
	}
	return reg, nil
}

func (g Generator) GetName() string {
	return GeneratorName
}

// Run validates the input from the flags and sets default values, after which
// it calls the custom client genrator to create wrappers. If there are any
// errors while generating interface wrappers, it prints it out.
func (g Generator) Run(ctx *genall.GenerationContext, f flag.Flags) error {
	if err := g.configure(f); err != nil {
		return err
	}

	return g.generate(ctx)
}

// configure sets the Generator's configuration using the given flags.
func (g *Generator) configure(f flag.Flags) error {
	if err := flag.ValidateFlags(f); err != nil {
		return err
	}

	absoluteInputDir, err := filepath.Abs(f.InputDir)
	if err != nil {
		return err
	}

	g.inputDir = absoluteInputDir
	g.baseInformersPackage = f.BaseInformersPackage
	g.outputDir = f.OutputDir

	g.headerText, err = util.GetHeaderText(f.GoHeaderFilePath)
	if err != nil {
		return err
	}

	gvs, err := clientgen.GetGV(f)
	if err != nil {
		return err
	}

	g.groupVersions = append(g.groupVersions, gvs...)

	return nil
}

// firstModule returns the Go Module path of the first non-nil *Package in the given list,
// otherwise it returns an empty string.
func firstModule(pkgs []*loader.Package) string {
	for _, pkg := range pkgs {
		if m := pkg.Module; m != nil {
			// Be greedy; take the first module path we find.
			return m.Path
		}
	}

	return ""
}

type sources struct {
	module string
	apis   map[types.Group][]informergen.API
}

func (g *Generator) sources(ctx *genall.GenerationContext) (*sources, error) {
	var errs []error
	srcs := &sources{
		apis: map[types.Group][]informergen.API{},
	}

	for _, group := range g.groups() {
		for _, version := range g.versionsFor(group) {
			path := filepath.Join(g.InputDir, group, version)
			pkgs, err := loader.LoadRootsWithConfig(&packages.Config{
				Mode: packages.NeedModule | packages.NeedTypesInfo,
			}, path)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to get packages for: %s/%s", group, version))
				continue
			}

			if srcs.module == "" {
				// All group-versions should be in the same module, so we can memoize the result for all source APIs
				if srcs.module = firstModule(pkgs); srcs.module == "" {
					errs = append(errs, fmt.Errorf("failed to find go module for: %s/%s", group, version))
					continue
				}
			}

			// Assign the pkgs obtained from loading roots to generation context.
			// TODO: Figure out if controller-tools generation runtime can be used to
			// wire in instead.
			ctx.Roots = pkgs

			for _, pkg := range pkgs {
				err = markers.EachType(ctx.Collector, pkg, func(info *markers.TypeInfo) {
					if !clientgen.IsEnabledForMethod(info) {
						// Skip types that don't have markers for generating clients
						return
					}

					srcs.apis[group] = append(sources.apis[group], informergen.API{
						Group:   group,
						Version: version,
						Kind:    info.Name,
					})

				})

				if err != nil {
					errs = append(errs, err)
				}
			}
		}
	}

	return srcs, loader.MaybeErrList(errs)
}

// generate first generates the wrapper for all the interfaces provided in the input.
// Then for each type defined in the input, it recursively wraps the subsequent
// interfaces to be kcp-aware.
func (g *Generator) generate(ctx *genall.GenerationContext) error {
	// if err := g.writeFactory(ctx); err != nil {
	// 	return err
	// }

	// if err := g.writeGeneric(ctx); err != nil {
	// 	return err
	// }

	var errs []error

	srcs, err := g.sources()
	if err != nil {
		errs = append(errs, err)
	}

	// TODO(kcp-dev): This will cause problems whenever listers don't exist at the expected package path.
	// Consider something more robust; e.g. a user provided command line argument.
	baseListersPackage := filepath.Join(srcs.module, g.outputDir, "listers")
	fmt.Printf("\tlisters: %s\n", baseListersPackage)

	for group, versions := range srcs.apis {
		for _, version := range versions {
			for _, api := range kinds {
				// TODO(kcp): Pipe io.Writers to format and write to disk
				var out bytes.Buffer

				// TODO(kcp): Pipe io.Writers to prepend header
				if err := g.writeHeader(&out); err != nil {
					errs = append(errs, err)
					continue
				}

				informergen.Informer{
					InformerPackage: filepath.Join(g.baseInformersPackage, group, version),
					ListerPackage:   filepath.Join(baseListersPackage, group, version),
					API:             api,
				}
				t := informergen.NewInformer(
					&out,
					version,
					g.baseInformersPackage,
					group,
					version,
					info.Name,
				)
			}

		}
	}

	for _, group := range g.groups() {
		versions := g.versionsFor(group)
		for _, version := range versions {
			var out bytes.Buffer
			// Assign the pkgs obtained from loading roots to generation context.
			// TODO: Figure out if controller-tools generation runtime can be used to
			// wire in instead.
			ctx.Roots = pkgs

			var collectedAPIs []string
			for _, root := range pkgs {
				if typeErr := markers.EachType(ctx.Collector, root, func(info *markers.TypeInfo) {
					var out bytes.Buffer

					// if not enabled for this type, skip
					if !clientgen.IsEnabledForMethod(info) {
						return
					}

					if err := g.writeHeader(&out); err != nil {
						root.AddError(err)
						return
					}

					t := informergen.NewInformer(
						&out,
						version,
						g.baseInformersPackage,
						baseListersPackage,
						group,
						version,
						info.Name,
					)

					if err := t.WriteContent(); err != nil {
						root.AddError(err)
						return
					}

					formatted, err := format.Source(out.Bytes())
					if err != nil {
						root.AddError(err)
						return
					}

					err = util.WriteContent(formatted, fmt.Sprintf("%ss.go", strings.ToLower(info.Name)), filepath.Join(g.outputDir, "informers", typedPackageName, group, version))
					if err != nil {
						root.AddError(err)
						return
					}

					collectedAPIs = append(collectedAPIs, info.Name)
				}); typeErr != nil {
					return typeErr
				}
			}

			var out bytes.Buffer
			if err := g.writeHeader(&out); err != nil {
				return err
			}

			t, err := informergen.NewVersionInterface(&out, version, g.baseInformersPackage, group, version, collectedAPIs)
			if err != nil {
				return err
			}

			if err := t.WriteContent(); err != nil {
				return err
			}

			formatted, err := format.Source(out.Bytes())
			if err != nil {
				return err
			}
		}

		if err := g.writeGroupInterface(ctx, group.String(), versions); err != nil {
			errs = append(errs, err)
			continue
		}
	}

	return loader.MaybeErrList(errs)
}

func (g *Generator) writeHeader(out io.Writer) error {
	n, err := out.Write([]byte(g.headerText))
	if err != nil {
		return err
	}

	if n < len([]byte(g.headerText)) {
		return errors.New("header text was not written properly.")
	}
	return nil
}

func (g *Generator) groups() (ret []types.Group) {
	groups := map[string]types.Group{}
	for _, gv := range g.groupVersions {
		groups[gv.Group.String()] = gv.Group
	}
	for _, group := range groups {
		ret = append(ret, group)
	}
	return
}

func (g *Generator) versionsFor(group types.Group) (versions []string) {
	visited := map[string]struct{}{}
	for _, gv := range g.groupVersions {
		if gv.Group != group || len(gv.Versions) < 1 {
			// Discard.
			continue
		}

		version := gv.Versions[0].Version.String()
		if _, ok := visited[version]; ok {
			// We've already visited this version.
			continue
		}

		versions = append(versions, version)
		visited[version] = struct{}{}
	}

	return
}

func (g *Generator) writeFactory(ctx *genall.GenerationContext) error {
	var out bytes.Buffer

	if err := g.writeHeader(&out); err != nil {
		return err
	}

	// TODO needs to know about each group
	t, err := informergen.NewFactory(&out, "externalversions")
	if err != nil {
		return err
	}
	t.WriteContent()

	outBytes := out.Bytes()
	formattedBytes, err := format.Source(outBytes)
	if err != nil {
		return err
	} else {
		outBytes = formattedBytes
	}

	return util.WriteContent(outBytes, "factory.go", filepath.Join(g.outputDir, "informers", typedPackageName))
}

func (g *Generator) writeGeneric(ctx *genall.GenerationContext) error {
	var out bytes.Buffer

	if err := g.writeHeader(&out); err != nil {
		return err
	}

	//TODO needs to know about each gvk

	t, err := informergen.NewGeneric(&out, "externalversions")
	if err != nil {
		return err
	}

	if err := t.WriteContent(); err != nil {
		return err
	}

	formatted, err := format.Source(out.Bytes())
	if err != nil {
		return err
	}

	return util.WriteContent(formatted, "generic.go", filepath.Join(g.outputDir, "informers", typedPackageName))
}

func (g *Generator) writeGroupInterface(ctx *genall.GenerationContext, group string, versions []string) error {
	var out bytes.Buffer
	if err := g.writeHeader(&out); err != nil {
		return err
	}

	t := informergen.NewGroupInterface(&out, group, g.baseInformersPackage, group, versions)
	if err := t.WriteContent(); err != nil {
		return err
	}

	formatted, err := format.Source(out.Bytes())
	if err != nil {
		return err
	}

	return util.WriteContent(formatted, "interface.go", filepath.Join(g.outputDir, "informers", typedPackageName, group))
}

func (g *Generator) writeGroupVersion(ctx *genall.GenerationContext, group, version string) error {
	abs, err := filepath.Abs(g.inputDir)
	if err != nil {
		return err
	}
	path := filepath.Join(abs, group, version)
	pkgs, err := loader.LoadRootsWithConfig(&packages.Config{
		Mode: packages.NeedModule | packages.NeedTypesInfo,
	}, path)
	if err != nil {
		return err
	}

	fmt.Printf("\tpackages: %v\n", pkgs)

	// TODO: Find inputDir module separately.
	var module string
	for _, pkg := range pkgs {
		fmt.Printf("\tpackage: %s\n", pkg.PkgPath)
		if m := pkg.Module; m != nil {
			// Be greedy; take the first module path we find.
			module = m.Path
			break
		}
	}

	fmt.Printf("\tmodule: %s\n", module)

	// TODO(kcp-dev): This will cause problems whenever listers don't exist at the expected package path.
	// Consider something more robust; e.g. a user provided command line argument.
	baseListersPackage := filepath.Join(module, g.outputDir, "listers")
	fmt.Printf("\tlisters: %s\n", baseListersPackage)

	// Assign the pkgs obtained from loading roots to generation context.
	// TODO: Figure out if controller-tools generation runtime can be used to
	// wire in instead.
	ctx.Roots = pkgs

	var collectedAPIs []string
	for _, root := range pkgs {
		if typeErr := markers.EachType(ctx.Collector, root, func(info *markers.TypeInfo) {
			var out bytes.Buffer

			// if not enabled for this type, skip
			if !clientgen.IsEnabledForMethod(info) {
				return
			}

			if err := g.writeHeader(&out); err != nil {
				root.AddError(err)
				return
			}

			t := informergen.NewInformer(
				&out,
				version,
				g.baseInformersPackage,
				baseListersPackage,
				group,
				version,
				info.Name,
			)

			if err := t.WriteContent(); err != nil {
				root.AddError(err)
				return
			}

			formatted, err := format.Source(out.Bytes())
			if err != nil {
				root.AddError(err)
				return
			}

			err = util.WriteContent(formatted, fmt.Sprintf("%ss.go", strings.ToLower(info.Name)), filepath.Join(g.outputDir, "informers", typedPackageName, group, version))
			if err != nil {
				root.AddError(err)
				return
			}

			collectedAPIs = append(collectedAPIs, info.Name)
		}); typeErr != nil {
			return typeErr
		}
	}

	var out bytes.Buffer
	if err := g.writeHeader(&out); err != nil {
		return err
	}

	t, err := informergen.NewVersionInterface(&out, version, g.baseInformersPackage, group, version, collectedAPIs)
	if err != nil {
		return err
	}

	if err := t.WriteContent(); err != nil {
		return err
	}

	formatted, err := format.Source(out.Bytes())
	if err != nil {
		return err
	}

	return util.WriteContent(formatted, "interface.go", filepath.Join(g.outputDir, "informers", typedPackageName, group, version))
}
