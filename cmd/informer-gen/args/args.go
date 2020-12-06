/*
Copyright 2017 The Kubernetes Authors.

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

package args

import (
	"fmt"
	"path"

	"github.com/spf13/pflag"
	"k8s.io/gengo/args"

	"k8s.io/code-generator/cmd/client-gen/types"
	codegenutil "k8s.io/code-generator/pkg/util"
)

// CustomArgs is used by the gengo framework to pass args specific to this generator.
type CustomArgs struct {
	// A sorted list of group versions to generate. For each of them the package path is found
	// in GroupVersionToInputPath.
	Groups []types.GroupVersions

	// Overrides for which types should be included in the client.
	IncludedTypesOverrides map[types.GroupVersion][]string

	VersionedClientSetPackage string
	InternalClientSetPackage  string
	ListersPackage            string
	SingleDirectory           bool

	// PluralExceptions define a list of pluralizer exceptions in Type:PluralType format.
	// The default list is "Endpoints:Endpoints"
	PluralExceptions []string
}

// NewDefaults returns default arguments for the generator.
func NewDefaults() (*args.GeneratorArgs, *CustomArgs) {
	genericArgs := args.Default().WithoutDefaultFlagParsing()
	customArgs := &CustomArgs{
		SingleDirectory:  false,
		PluralExceptions: []string{"Endpoints:Endpoints"},
	}
	genericArgs.CustomArgs = customArgs

	if pkg := codegenutil.CurrentPackage(); len(pkg) != 0 {
		genericArgs.OutputPackagePath = path.Join(pkg, "pkg/client/informers")
		customArgs.VersionedClientSetPackage = path.Join(pkg, "pkg/client/clientset/versioned")
		customArgs.InternalClientSetPackage = path.Join(pkg, "pkg/client/clientset/internalversion")
		customArgs.ListersPackage = path.Join(pkg, "pkg/client/listers")
	}

	return genericArgs, customArgs
}

// AddFlags add the generator flags to the flag set.
func (ca *CustomArgs) AddFlags(fs *pflag.FlagSet, inputBase string) {
	gvsBuilder := NewGroupVersionsBuilder(&ca.Groups)
	pflag.Var(NewGVPackagesValue(gvsBuilder, nil), "input", "group/versions that client-gen will generate clients for. At most one version per group is allowed. Specified in the format \"group1/version1,group2/version2...\".")
	pflag.Var(NewGVTypesValue(&ca.IncludedTypesOverrides, []string{}), "included-types-overrides", "list of group/version/type for which client should be generated. By default, client is generated for all types which have genclient in types.go. This overrides that. For each groupVersion in this list, only the types mentioned here will be included. The default check of genclient will be used for other group versions.")
	pflag.Var(NewInputBasePathValue(gvsBuilder, inputBase), "input-base", "base path to look for the api group.")

	fs.StringVar(&ca.InternalClientSetPackage, "internal-clientset-package", ca.InternalClientSetPackage, "the full package name for the internal clientset to use")
	fs.StringVar(&ca.VersionedClientSetPackage, "versioned-clientset-package", ca.VersionedClientSetPackage, "the full package name for the versioned clientset to use")
	fs.StringVar(&ca.ListersPackage, "listers-package", ca.ListersPackage, "the full package name for the listers to use")
	fs.BoolVar(&ca.SingleDirectory, "single-directory", ca.SingleDirectory, "if true, omit the intermediate \"internalversion\" and \"externalversions\" subdirectories")
	fs.StringSliceVar(&ca.PluralExceptions, "plural-exceptions", ca.PluralExceptions, "list of comma separated plural exception definitions in Type:PluralizedType format")
}

// Validate checks the given arguments.
func Validate(genericArgs *args.GeneratorArgs) error {
	customArgs := genericArgs.CustomArgs.(*CustomArgs)

	if len(genericArgs.OutputPackagePath) == 0 {
		return fmt.Errorf("output package cannot be empty")
	}
	if len(customArgs.VersionedClientSetPackage) == 0 {
		return fmt.Errorf("versioned clientset package cannot be empty")
	}
	if len(customArgs.ListersPackage) == 0 {
		return fmt.Errorf("listers package cannot be empty")
	}

	return nil
}

// GroupVersionPackages returns a map from GroupVersion to the package with the types.go.
func (ca *CustomArgs) GroupVersionPackages() map[types.GroupVersion]string {
	res := map[types.GroupVersion]string{}
	for _, pkg := range ca.Groups {
		for _, v := range pkg.Versions {
			res[types.GroupVersion{Group: pkg.Group, Version: v.Version}] = v.Package
		}
	}
	return res
}
