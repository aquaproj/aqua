package resolve

import (
	"fmt"
	"log/slog"
	"maps"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/asset"
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
	rts, err = addEmulated(pkgInfo, rts, supportedEnvs)
	if err != nil {
		return nil, err
	}
	return ExpandByVariants(pkgInfo, rts), nil
}

// addEmulated adds the arm64 environments that run the amd64 build.
//
// supported_envs can name windows/amd64 alone while windows_arm_emulation says a
// Windows arm64 machine runs that build, and aqua installs it there: the check at
// install time reads the two together. Reading supported_envs alone left the arm64
// entry out, so a release publishing only an amd64 Windows build, such as
// jacobdeichert/mask 0.11.5, had no entry for windows/arm64. Rosetta 2 is the same
// on macOS.
func addEmulated(pkgInfo *registry.PackageInfo, rts []*runtime.Runtime, supportedEnvs []string) ([]*runtime.Runtime, error) {
	have := envSet(rts)
	var allowed map[string]struct{}
	if len(supportedEnvs) != 0 {
		cfgRTs, err := runtime.GetRuntimesFromEnvs(supportedEnvs)
		if err != nil {
			return nil, fmt.Errorf("get the supported environments: %w", err)
		}
		allowed = envSet(cfgRTs)
	}
	for _, e := range []struct {
		goos     string
		emulated bool
	}{
		{osDarwin, pkgInfo.Rosetta2},
		{osWindows, pkgInfo.WindowsARMEmulation},
	} {
		if !e.emulated {
			continue
		}
		if _, ok := have[e.goos+"/amd64"]; !ok {
			continue
		}
		to := e.goos + "/arm64"
		if _, ok := have[to]; ok {
			continue
		}
		if allowed != nil {
			if _, ok := allowed[to]; !ok {
				continue
			}
		}
		rts = append(rts, &runtime.Runtime{GOOS: e.goos, GOARCH: archArm64})
	}
	return rts, nil
}

func envSet(rts []*runtime.Runtime) map[string]struct{} {
	set := make(map[string]struct{}, len(rts))
	for _, rt := range rts {
		set[rt.Env()] = struct{}{}
	}
	return set
}

const (
	osDarwin  = "darwin"
	osWindows = "windows"
	archArm64 = "arm64"
)

// emulationOf says whether the entry runs the amd64 build on an arm64 machine.
//
// It is read off what was resolved rather than off the definition, since a definition can
// say rosetta2 and still resolve to a universal build, a native one named in an override,
// or a go_install path, none of which would stop working without Rosetta 2. The entry
// depends on emulation when, with emulation turned off,
//
//   - the asset is another one: the asset was named from the amd64 architecture,
//   - the environment isn't supported at all: it is there only because of emulation, or
//   - the asset is the same one and its name says amd64 and not arm64 or universal:
//     Songmu/gocredits writes gocredits_{{.Version}}_{{.OS}}_amd64.{{.Format}}, the
//     same name whichever architecture asks.
//
// Only the emulation a definition declares is recorded, for the OS it is declared for:
// rosetta2 on macOS, windows_arm_emulation on Windows. A definition that sends arm64 to
// the amd64 build through replacements alone -- snaplet/cli's arm64: x64 -- isn't.
//
// It returns whether that is through Rosetta 2 and whether through Windows' emulation.
func emulationOf(param *Param, info *registry.PackageInfo, rt *runtime.Runtime, assetName, url string) (bool, bool, error) {
	emulated, err := runsAmd64Build(param, info, rt, assetName, url)
	if err != nil || !emulated {
		return false, false, err
	}
	return rt.GOOS == osDarwin, rt.GOOS == osWindows, nil
}

func runsAmd64Build(param *Param, info *registry.PackageInfo, rt *runtime.Runtime, assetName, url string) (bool, error) {
	if !declaresEmulation(info, rt) {
		return false, nil
	}
	native := info.Copy()
	native.Rosetta2 = false
	native.WindowsARMEmulation = false
	pkg := &config.Package{
		Package:     &aqua.Package{Name: param.PkgName, Version: param.Version},
		PackageInfo: native,
	}
	if err := fillVars(pkg); err != nil {
		return false, err
	}
	nativeAsset, nativeURL, err := renderLocation(pkg, native, rt)
	if err != nil {
		return false, fmt.Errorf("resolve without emulation: %w", err)
	}
	return nativeAsset != assetName || nativeURL != url || amd64Only(native, rt, assetName, url), nil
}

// declaresEmulation says whether the definition declares that an arm64 machine of the
// environment's OS runs the amd64 build.
func declaresEmulation(info *registry.PackageInfo, rt *runtime.Runtime) bool {
	if rt.GOARCH != archArm64 {
		return false
	}
	return rt.GOOS == osDarwin && info.Rosetta2 || rt.GOOS == osWindows && info.WindowsARMEmulation
}

// amd64Only says whether an asset that is the same with or without emulation is the amd64
// build: the environment isn't supported without emulation, or the name says amd64.
func amd64Only(native *registry.PackageInfo, rt *runtime.Runtime, assetName, url string) bool {
	if !native.CheckSupportedEnvs(rt.GOOS, rt.GOARCH, rt.Env()) {
		return true
	}
	name := assetName
	if name == "" {
		name = path.Base(url)
	}
	return namesAmd64(name)
}

// namesAmd64 says whether an asset's name says it is an amd64 build and not an arm64 or
// universal one. The spellings have to be whole words, so that a tool named x64tool or a
// repository named after arm64 doesn't decide it.
func namesAmd64(name string) bool {
	s := strings.ToLower(name)
	return amd64Word.MatchString(s) && !nativeWord.MatchString(s)
}

var (
	amd64Word  = regexp.MustCompile(`(^|[^a-z0-9])(amd64|x86[-_]64|x64|64-?bit)([^a-z0-9]|$)`)
	nativeWord = regexp.MustCompile(`(^|[^a-z0-9])(arm64|aarch64|universal)([^a-z0-9]|$)`)
)

// nothingToInstall says a package resolves to neither an asset nor a URL on the
// environment. go_install and cargo are identified by their path and crate instead, so
// they never are.
func nothingToInstall(info *registry.PackageInfo, assetName, url string) bool {
	return assetName == "" && url == "" && info.Type != registry.PkgInfoTypeGoInstall && info.Type != registry.PkgInfoTypeCargo
}

// renderLocation resolves where the package is downloaded from on the environment: the
// asset name, and the URL for an http package.
func renderLocation(pkg *config.Package, info *registry.PackageInfo, rt *runtime.Runtime) (string, string, error) {
	assetName, err := pkg.RenderAsset(rt)
	if err != nil {
		return "", "", fmt.Errorf("render the asset name for %s: %w", rt.Env(), err)
	}
	url, err := renderURL(pkg, info, rt)
	if err != nil {
		return "", "", err
	}
	return assetName, url, nil
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
	return resolveInfo(param, forRuntime(versioned, filesFrom, rt), rt)
}

// ResolveSelected resolves the entry of an asset already chosen for its environment.
//
// The asset is what the release itself says, picked by its name rather than rendered
// from the definition, so the definition's asset, format and the spellings it would
// render them with don't apply: the entry names the asset as the release lists it. What
// the definition still supplies is what a release can't say, such as the executables
// inside the archive and how the asset is signed, and those are rendered with the
// spellings the asset's own name uses, so that a signature named after the asset's
// platform names the release's file.
func ResolveSelected(logger *slog.Logger, param *Param, sel *asset.Selected) (*lockfile.Package, error) {
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
	rt := &runtime.Runtime{GOOS: sel.OS, GOARCH: sel.Arch, LibC: sel.Libc}
	info := forRuntime(versioned, filesFrom, rt)
	info.Asset = sel.Name
	info.Format = sel.Info.Format
	info.Replacements = withSpellings(info.Replacements, sel.Info)
	no := false
	info.AppendExt = &no
	info.CompleteWindowsExt = &no
	// An emulated environment installs the amd64 build, and a template naming the
	// architecture has to name the build's.
	info.Rosetta2 = sel.Emulated && sel.OS == "darwin"
	info.WindowsARMEmulation = sel.Emulated && sel.OS == "windows"
	return resolveInfo(param, info, rt)
}

// withSpellings is the definition's replacements with the asset's own spellings over them.
//
// The OS and the architecture of the asset are spelled as its name spells them, which is
// the platform's own name when the name uses that: a definition writing darwin as macOS
// for its other assets must not rename it in a template rendered for an asset that says
// darwin. An architecture the name doesn't say is left to the definition.
func withSpellings(defined registry.Replacements, info *asset.AssetInfo) registry.Replacements {
	out := make(registry.Replacements, len(defined)+2) //nolint:mnd
	maps.Copy(out, defined)
	out[info.OS] = spelled(info.Replacements, info.OS)
	if !info.NoArch {
		out[info.Arch] = spelled(info.Replacements, info.Arch)
	}
	return out
}

// spelled is how the name spells key, which is key itself when it isn't replaced.
func spelled(replacements map[string]string, key string) string {
	if v, ok := replacements[key]; ok {
		return v
	}
	return key
}

// resolveInfo resolves the definition as it applies to one environment.
func resolveInfo(param *Param, info *registry.PackageInfo, rt *runtime.Runtime) (*lockfile.Package, error) {
	pkg := &config.Package{
		Package:     &aqua.Package{Name: param.PkgName, Version: param.Version},
		PackageInfo: info,
	}
	if err := fillVars(pkg); err != nil {
		return nil, err
	}

	assetName, url, err := renderLocation(pkg, info, rt)
	if err != nil {
		return nil, err
	}
	// A package that resolves to neither an asset nor a URL has nothing to install
	// on this environment. go_install and cargo are identified by their path and
	// crate instead, so they are not caught by this.
	if nothingToInstall(info, assetName, url) {
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
	rosetta2, windowsARMEmulation, err := emulationOf(param, info, rt, assetName, url)
	if err != nil {
		return nil, err
	}

	return &lockfile.Package{
		Name:                param.PkgName,
		Version:             param.Version,
		OS:                  rt.GOOS,
		Arch:                rt.GOARCH,
		Variants:            variantsOf(rt),
		Rosetta2:            rosetta2,
		WindowsARMEmulation: windowsARMEmulation,
		Type:                info.Type,
		RepoOwner:           info.RepoOwner,
		RepoName:            info.RepoName,
		// The instance, for the types that are on one. Without it the entry would
		// name a repository on github.com, where the same owner and name are
		// somebody else's. A definition may leave it to the type's own default --
		// a gitlab_release package is on gitlab.com unless it says otherwise -- and
		// the entry says which instance either way, because what reads the entry
		// reads nothing else.
		Host:                       info.GetHost(),
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
		Packslip:                   sign.packslip,
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
