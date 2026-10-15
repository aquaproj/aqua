package lockfile

import (
	"cmp"
	"slices"
	"strings"
)

// Sort orders the file deterministically.
//
// The entries are a flat list, so the same set of packages has to come out in the
// same order every time or a regenerated lock file produces a diff that says
// nothing. Nothing in the format depends on the order.
func (lf *LockFile) Sort() {
	slices.SortFunc(lf.Packages, comparePackages)
	for _, pkg := range lf.Packages {
		slices.SortFunc(pkg.Files, compareFiles)
	}
}

func comparePackages(a, b *Package) int {
	if d := cmp.Compare(a.Name, b.Name); d != 0 {
		return d
	}
	if d := cmp.Compare(a.Version, b.Version); d != 0 {
		return d
	}
	if d := cmp.Compare(a.OS, b.OS); d != 0 {
		return d
	}
	if d := cmp.Compare(a.Arch, b.Arch); d != 0 {
		return d
	}
	// Two entries can share all of the above and differ by variant, such as a
	// linux/amd64 build for musl and one for glibc.
	return cmp.Compare(variantKey(a), variantKey(b))
}

func compareFiles(a, b *File) int {
	if d := cmp.Compare(a.Name, b.Name); d != 0 {
		return d
	}
	return cmp.Compare(a.Src, b.Src)
}

// variantKey renders the variants as a string that sorts stably.
func variantKey(pkg *Package) string {
	if len(pkg.Variants) == 0 {
		return ""
	}
	keys := make([]string, 0, len(pkg.Variants))
	for k := range pkg.Variants {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k + "=" + pkg.Variants[k] + ",")
	}
	return b.String()
}

// Has reports whether the lock file holds the package at that version.
//
// It answers for the package as a whole rather than for one environment: a lock
// file carries every environment a package supports, so a version that is present
// at all is present for all of them.
func (lf *LockFile) Has(name, version string) bool {
	for _, pkg := range lf.Packages {
		if pkg.Name == name && pkg.Version == version {
			return true
		}
	}
	return false
}

// Retain drops the entries a predicate doesn't declare, and reports how many went.
//
// What a lock file is for is the packages a configuration asks for, so an entry for one
// it no longer asks for answers a question nobody puts. Nothing looks one up -- a lookup
// is by name and version -- so what it costs is the file saying more than it means.
func (lf *LockFile) Retain(declared func(name, version string) bool) int {
	before := len(lf.Packages)
	lf.Packages = slices.DeleteFunc(lf.Packages, func(pkg *Package) bool {
		// A null entry is nothing a configuration asks for either, and it is the one
		// thing in the file that can't be looked up at all.
		return pkg == nil || !declared(pkg.Name, pkg.Version)
	})
	return before - len(lf.Packages)
}

// Remove drops every entry of the package at that version.
func (lf *LockFile) Remove(name, version string) {
	lf.Packages = slices.DeleteFunc(lf.Packages, func(pkg *Package) bool {
		return pkg.Name == name && pkg.Version == version
	})
}
