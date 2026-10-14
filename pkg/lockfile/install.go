package lockfile

import (
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

// Find returns the entry describing name at version on rt, or nil when the lock file
// has none.
//
// Several entries can share an os and arch and differ only by variants, such as a
// musl build beside a glibc one. The most specific match wins, the same way an
// Override does: an entry constraining libc beats one constraining nothing, so the
// unconstrained entry acts as the fallback rather than shadowing its siblings.
//
// When no entry matches, a statically linked build is installed: it needs no libc, so
// it runs whichever one the machine has, including one whose libc can't be told. A
// release publishing a glibc and a musl build has an entry for each and none for a
// machine that is neither, and the musl build is usually the static one.
func (lf *LockFile) Find(name, version string, rt *runtime.Runtime) *Package {
	if lf == nil {
		return nil
	}
	var best, static *Package
	for _, pkg := range lf.Packages {
		if !pkg.isFor(name, version, rt) {
			continue
		}
		if static == nil && pkg.isStatic() {
			static = pkg
		}
		if !matchVariants(pkg.Variants, rt) {
			continue
		}
		if best == nil || len(pkg.Variants) > len(best.Variants) {
			best = pkg
		}
	}
	if best == nil {
		return static
	}
	return best
}

// isFor says whether the entry describes name at version on rt's os and arch, whatever its
// variants say.
func (p *Package) isFor(name, version string, rt *runtime.Runtime) bool {
	return p.Name == name && p.Version == version && p.OS == rt.GOOS && p.Arch == rt.GOARCH
}

// isStatic says whether the entry is a build needing no libc, for a machine aqua can tell
// it is for.
func (p *Package) isStatic() bool {
	return p.LinkedLibc == linkedStatic && supportedVariants(p.Variants)
}

// linkedStatic is the LinkedLibc of a build that needs no libc.
const linkedStatic = "static"

// supportedVariants says whether aqua evaluates every key of the variants. An entry
// selecting on something aqua can't tell is for a machine it can't say it is.
func supportedVariants(variants map[string]string) bool {
	for key := range variants {
		if !registry.IsSupportedVariantKey(key) {
			return false
		}
	}
	return true
}

func matchVariants(variants map[string]string, rt *runtime.Runtime) bool {
	for key, value := range variants {
		// A key aqua doesn't evaluate can't be shown to hold, so the entry is not a
		// match. Installing it anyway would mean guessing.
		if !registry.IsSupportedVariantKey(key) {
			return false
		}
		if runtimeVariantValue(rt, key) != value {
			return false
		}
	}
	return true
}

func runtimeVariantValue(rt *runtime.Runtime, key string) string {
	if key == "libc" {
		return rt.LibC
	}
	return ""
}

// PackageInfo returns the entry as the package definition aqua installs from.
//
// Nothing here is a template: the entry was resolved for one environment when the lock
// file was written. append_ext and complete_windows_ext are therefore turned off,
// since both exist to finish a name a registry left incomplete and would edit a name
// that is already final.
func (p *Package) PackageInfo() *registry.PackageInfo {
	no := false
	return &registry.PackageInfo{
		Name:                       p.Name,
		Type:                       p.Type,
		RepoOwner:                  p.RepoOwner,
		RepoName:                   p.RepoName,
		Host:                       p.Host,
		Asset:                      p.Asset,
		URL:                        p.URL,
		Format:                     p.Format,
		Path:                       p.Path,
		Crate:                      p.Crate,
		Cargo:                      p.Cargo,
		Files:                      p.registryFiles(),
		AppendExt:                  &no,
		CompleteWindowsExt:         &no,
		Private:                    p.Private,
		Cosign:                     p.Cosign,
		GitHubArtifactAttestations: p.GitHubArtifactAttestations,
		Minisign:                   p.Minisign,
		Packslip:                   p.Packslip,
		SLSAProvenance:             p.SLSAProvenance,
	}
}

func (p *Package) registryFiles() []*registry.File {
	if len(p.Files) == 0 {
		return nil
	}
	files := make([]*registry.File, len(p.Files))
	for i, f := range p.Files {
		files[i] = &registry.File{
			Name: f.Name,
			Src:  f.Src,
		}
	}
	return files
}

// NeedsChecksum reports whether the entry must carry a checksum.
//
// go_install and cargo build through another tool, which resolves and verifies what
// it fetches for itself; aqua never downloads an artifact for them, so there is
// nothing for it to check. Every other type downloads something, and a lock file
// entry without a checksum for it would be a hole in the guarantee the file exists
// to make.
func (p *Package) NeedsChecksum() bool {
	switch p.Type {
	case registry.PkgInfoTypeGoInstall, registry.PkgInfoTypeCargo:
		return false
	default:
		return true
	}
}
