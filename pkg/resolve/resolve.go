package resolve

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

// Param is what to resolve.
type Param struct {
	// PkgName and Version identify the package. They are copied onto every entry,
	// since an entry has to say what it describes.
	PkgName string
	Version string
	// PkgInfo is the registry's definition of the package.
	PkgInfo *registry.PackageInfo
	// SupportedEnvs narrows which environments to resolve, the way aqua.yaml's
	// checksum.supported_envs does. Empty means every environment the package
	// itself supports.
	SupportedEnvs []string
	// FilesFrom, when set, supplies the executables instead of PkgInfo.
	//
	// It exists for a caller that infers a package's asset naming from the release
	// itself, so that an upstream renaming can't break it, because a release says
	// nothing about what the executables inside the archive are called. Those come
	// from the written definition, resolved for the same environment, since an
	// override can move them: cli/cli's Windows build puts the executable at
	// bin/gh.exe while every other platform has it under a versioned directory.
	FilesFrom *registry.PackageInfo
}

// Resolve returns one entry per environment the package supports.
//
// Everything is resolved but the checksum: the asset name, the download URL, the
// format and the path of each executable inside the archive. What comes back is what
// a lock file records and what aqua-registry-g2 serves, which is the same thing
// described twice, so it is produced once here.
func Resolve(logger *slog.Logger, param *Param) ([]*lockfile.Package, error) {
	// The version is resolved before the environments are, because a
	// version_override carries its own supported_envs and overrides: which
	// environments exist, and which variants they have, is a property of the version
	// rather than of the package.
	versioned, err := param.PkgInfo.SetVersion(logger, param.Version)
	if err != nil {
		return nil, fmt.Errorf("evaluate the version constraints: %w", err)
	}
	var filesFrom *registry.PackageInfo
	if param.FilesFrom != nil {
		filesFrom, err = param.FilesFrom.SetVersion(logger, param.Version)
		if err != nil {
			return nil, fmt.Errorf("evaluate the version constraints of the files: %w", err)
		}
	}
	rts, err := Runtimes(versioned, param.SupportedEnvs)
	if err != nil {
		return nil, err
	}
	pkgs := make([]*lockfile.Package, 0, len(rts))
	for _, rt := range rts {
		pkg, err := resolveOne(param, versioned, filesFrom, rt)
		if err != nil {
			return nil, err
		}
		if pkg == nil {
			continue
		}
		pkgs = append(pkgs, pkg)
	}
	if len(pkgs) == 0 {
		return nil, errNoSupportedEnv
	}
	return dropRedundantVariants(pkgs), nil
}

// dropRedundantVariants removes entries a machine would resolve the same way without
// them.
//
// A package that declares a glibc override changing nothing produces a glibc entry
// identical to the unconstrained one. Keeping both writes the same answer twice: a
// glibc machine that finds no glibc entry falls through to the unconstrained entry
// and installs exactly the same thing. The unconstrained one is what stays, because
// it also answers for the machines whose libc is something else or can't be detected.
func dropRedundantVariants(pkgs []*lockfile.Package) []*lockfile.Package {
	fallbacks := make(map[string]*lockfile.Package, len(pkgs))
	for _, pkg := range pkgs {
		if len(pkg.Variants) == 0 {
			fallbacks[pkg.OS+"/"+pkg.Arch] = pkg
		}
	}
	out := make([]*lockfile.Package, 0, len(pkgs))
	for _, pkg := range pkgs {
		if fallback, ok := fallbacks[pkg.OS+"/"+pkg.Arch]; ok && len(pkg.Variants) != 0 && sameExceptVariants(fallback, pkg) {
			continue
		}
		out = append(out, pkg)
	}
	return out
}

func sameExceptVariants(a, b *lockfile.Package) bool {
	x := *a
	y := *b
	x.Variants = nil
	y.Variants = nil
	return reflect.DeepEqual(&x, &y)
}

// Runtimes returns the environments to resolve a package for.
func Runtimes(pkgInfo *registry.PackageInfo, supportedEnvs []string) ([]*runtime.Runtime, error) {
	rts, err := checksum.GetRuntimesFromSupportedEnvs(supportedEnvs, pkgInfo.SupportedEnvs)
	if err != nil {
		return nil, fmt.Errorf("get the supported environments: %w", err)
	}
	return ExpandByVariants(pkgInfo, rts), nil
}

// forRuntime is the definition as it applies to one environment.
//
// The copy comes first because SetVersion returns the receiver itself when the
// package has no top-level version_constraint, and OverrideByRuntime then mutates it
// in place. Without the copy, an override applied for one environment leaks into
// every environment resolved afterwards.
func forRuntime(versioned, filesFrom *registry.PackageInfo, rt *runtime.Runtime) *registry.PackageInfo {
	info := versioned.Copy()
	info.OverrideByRuntime(rt)
	if filesFrom == nil {
		return info
	}
	other := filesFrom.Copy()
	other.OverrideByRuntime(rt)
	if files := other.GetFiles(); len(files) > 0 {
		info.Files = files
	}
	return info
}

// resolveOne resolves a single environment, or returns nil when the package doesn't
// support it or resolves to nothing there.
func resolveOne(param *Param, versioned, filesFrom *registry.PackageInfo, rt *runtime.Runtime) (*lockfile.Package, error) {
	info := forRuntime(versioned, filesFrom, rt)
	pkg := &config.Package{
		Package:     &aqua.Package{Name: param.PkgName, Version: param.Version},
		PackageInfo: info,
	}
	if err := fillVars(pkg); err != nil {
		return nil, err
	}

	assetName, err := pkg.RenderAsset(rt)
	if err != nil {
		return nil, fmt.Errorf("render the asset name for %s: %w", rt.Env(), err)
	}
	url, err := renderURL(pkg, info, rt)
	if err != nil {
		return nil, err
	}
	// A package that resolves to neither an asset nor a URL has nothing to install
	// on this environment. go_install and cargo are identified by their path and
	// crate instead, so they are not caught by this.
	if assetName == "" && url == "" && info.Type != registry.PkgInfoTypeGoInstall && info.Type != registry.PkgInfoTypeCargo {
		return nil, nil //nolint:nilnil
	}

	files, err := renderFiles(pkg, info, rt)
	if err != nil {
		return nil, err
	}
	sign, err := renderSigning(pkg, info, rt, assetName)
	if err != nil {
		return nil, err
	}

	return &lockfile.Package{
		Name:                       param.PkgName,
		Version:                    param.Version,
		OS:                         rt.GOOS,
		Arch:                       rt.GOARCH,
		Variants:                   variantsOf(rt),
		Type:                       info.Type,
		RepoOwner:                  info.RepoOwner,
		RepoName:                   info.RepoName,
		Asset:                      assetName,
		URL:                        url,
		Format:                     info.GetFormat(),
		Path:                       info.Path,
		Crate:                      info.Crate,
		Cargo:                      info.Cargo,
		Files:                      files,
		Private:                    info.Private,
		Cosign:                     sign.cosign,
		GitHubArtifactAttestations: info.GitHubArtifactAttestations,
		Minisign:                   sign.minisign,
		SLSAProvenance:             sign.slsa,
	}, nil
}

// renderURL resolves the download URL of an http package. Other types are identified
// by their asset name instead, so they have no URL to render.
func renderURL(pkg *config.Package, info *registry.PackageInfo, rt *runtime.Runtime) (string, error) {
	if info.Type != registry.PkgInfoTypeHTTP {
		return "", nil
	}
	url, err := pkg.RenderURL(rt)
	if err != nil {
		return "", fmt.Errorf("render the URL for %s: %w", rt.Env(), err)
	}
	return url, nil
}

// renderFiles resolves the templates in files[].src for the environment.
//
// aqua renders files[].src through an unexported method that supplies variables
// RenderTemplateString doesn't, such as the asset name without its extension, so
// rendering the template directly leaves "<no value>" in the result. ExePath is the
// only exported path that goes through it, so the source path is recovered by
// removing the package directory it prefixes.
func renderFiles(pkg *config.Package, info *registry.PackageInfo, rt *runtime.Runtime) ([]*lockfile.File, error) {
	infoFiles := info.GetFiles()
	if len(infoFiles) == 0 {
		return nil, nil
	}
	pkgPath, err := pkg.PkgPath(rt)
	if err != nil {
		return nil, fmt.Errorf("get the package path for %s: %w", rt.Env(), err)
	}
	files := make([]*lockfile.File, 0, len(infoFiles))
	for _, f := range infoFiles {
		exePath, err := pkg.ExePath("", f, rt)
		if err != nil {
			return nil, fmt.Errorf("render files[].src for %s: %w", rt.Env(), err)
		}
		files = append(files, &lockfile.File{
			Name: f.Name,
			Src:  strings.TrimPrefix(exePath, pkgPath+string(filepath.Separator)),
		})
	}
	return files, nil
}

// variantsOf records what tells an entry apart from a sibling sharing its os and
// arch. An environment with nothing to distinguish it carries no variants rather
// than an empty map, so that the two are not written differently.
func variantsOf(rt *runtime.Runtime) map[string]string {
	if rt.LibC == "" {
		return nil
	}
	return map[string]string{"libc": rt.LibC}
}

// fillVars gives every variable the definition declares a value, or says which one it
// could not.
//
// A definition may write part of its URL as a variable with a default, the way
// flutter/flutter names its release channel. Nothing here supplies one, so the
// defaults are all there is: without them the template renders "<no value>" into the
// URL and the download fails with a 404 that says nothing about why.
//
// ApplyVars leaves a variable that is neither required nor defaulted unset, which
// renders the same way. Where a person supplies the value that is tolerable, because
// they see the result; here there is nobody to ask, and the text goes into a file
// that is served as the answer. A definition whose variable nothing can fill cannot
// be resolved statically at all, so this names the variable instead.
func fillVars(pkg *config.Package) error {
	if err := pkg.ApplyVars(); err != nil {
		return fmt.Errorf("apply the package variables: %w", err)
	}
	for _, v := range pkg.PackageInfo.Vars {
		if _, ok := pkg.Package.Vars[v.Name]; !ok {
			return fmt.Errorf("%w: %s", errVarHasNoValue, v.Name)
		}
	}
	return nil
}
