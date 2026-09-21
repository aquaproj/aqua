// Package resolve turns a registry's definition of a package into entries resolved
// for each environment it supports.
//
// A registry describes a package with templates and overrides, which have to be
// evaluated per operating system, architecture and variant before anything can be
// installed or written to a lock file. Doing that in one place is what keeps aqua and
// the tool that generates aqua-registry-g2 from disagreeing about what a registry
// means.
package resolve

import (
	"maps"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

// ExpandByVariants returns rts with each entry potentially duplicated
// to cover every variant value combination declared in pkgInfo.Overrides for
// the base runtime's (GOOS, GOARCH). For packages without variant-aware
// overrides this is effectively a no-op and the input runtimes are returned.
//
// Without this expansion, update-checksum would only see whichever single
// Override [pkg/config/registry.Override.Match] picks first for a given
// runtime — silently dropping the assets of any sibling Override that differs
// only by variants (e.g. libc=musl vs libc=glibc).
func ExpandByVariants(pkgInfo *registry.PackageInfo, rts []*runtime.Runtime) []*runtime.Runtime {
	if len(pkgInfo.Overrides) == 0 {
		return rts
	}
	keys := registry.SupportedVariantKeys()
	expanded := make([]*runtime.Runtime, 0, len(rts))
	seen := map[string]struct{}{}
	add := func(rt *runtime.Runtime) {
		k := RuntimeKey(rt)
		if _, ok := seen[k]; ok {
			return
		}
		seen[k] = struct{}{}
		expanded = append(expanded, rt)
	}
	for _, rt := range rts {
		matching := platformCandidateOverrides(pkgInfo.Overrides, rt)
		if !hasVariantsOverride(matching) {
			add(rt)
			continue
		}
		valueSets := collectVariantValueSets(matching, keys)
		for _, combo := range cartesianProduct(valueSets) {
			newRT := *rt
			applyVariantCombo(&newRT, combo)
			add(&newRT)
		}
	}
	return expanded
}

// platformCandidateOverrides returns Overrides that match rt by GOOS / GOArch
// / Envs, ignoring Variants. Overrides referencing an unsupported variant key
// are dropped because [pkg/config/registry.Override.Match] would reject them
// for any runtime anyway.
func platformCandidateOverrides(overrides []*registry.Override, rt *runtime.Runtime) []*registry.Override {
	out := make([]*registry.Override, 0, len(overrides))
	for _, ov := range overrides {
		if !ov.MatchPlatform(rt) {
			continue
		}
		if !allVariantKeysSupported(ov) {
			continue
		}
		out = append(out, ov)
	}
	return out
}

func allVariantKeysSupported(ov *registry.Override) bool {
	for _, v := range ov.Variants {
		if !registry.IsSupportedVariantKey(v.Key) {
			return false
		}
	}
	return true
}

func hasVariantsOverride(overrides []*registry.Override) bool {
	for _, ov := range overrides {
		if len(ov.Variants) > 0 {
			return true
		}
	}
	return false
}

// collectVariantValueSets returns, for each supported variant key, the set of values
// to enumerate.
//
// The empty value is always among them. It stands for an environment that constrains
// nothing, which is reachable however the overrides are written: an Override matches
// only when the runtime's value equals its own, so a machine whose libc is something
// else, or whose libc couldn't be detected at all, falls through to the package's own
// definition. Leaving it out would drop the environment that uses the base asset,
// which for a package whose base is the musl build is the musl machines.
func collectVariantValueSets(overrides []*registry.Override, keys []string) map[string]map[string]struct{} {
	sets := make(map[string]map[string]struct{}, len(keys))
	for _, key := range keys {
		sets[key] = map[string]struct{}{"": {}}
	}
	for _, ov := range overrides {
		for _, v := range ov.Variants {
			if _, ok := sets[v.Key]; ok {
				sets[v.Key][v.Value] = struct{}{}
			}
		}
	}
	return sets
}

func cartesianProduct(sets map[string]map[string]struct{}) []map[string]string {
	combos := []map[string]string{{}}
	for key, values := range sets {
		next := make([]map[string]string, 0, len(combos)*len(values))
		for _, c := range combos {
			for v := range values {
				nc := make(map[string]string, len(c)+1)
				maps.Copy(nc, c)
				nc[key] = v
				next = append(next, nc)
			}
		}
		combos = next
	}
	return combos
}

func applyVariantCombo(rt *runtime.Runtime, combo map[string]string) {
	if v, ok := combo["libc"]; ok {
		rt.LibC = v
	}
}

// RuntimeKey returns a unique map key for rt that includes the variant axes, so that
// runtimes differing only by variant, such as musl and glibc, do not collide.
func RuntimeKey(rt *runtime.Runtime) string {
	return rt.GOOS + "/" + rt.GOARCH + "/" + rt.LibC
}
