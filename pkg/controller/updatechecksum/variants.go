package updatechecksum

import (
	"maps"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

// expandRuntimesByVariants returns rts with each entry potentially duplicated
// to cover every variant value combination declared in pkgInfo.Overrides for
// the base runtime's (GOOS, GOARCH). For packages without variant-aware
// overrides this is effectively a no-op and the input runtimes are returned.
//
// Without this expansion, update-checksum would only see whichever single
// Override [pkg/config/registry.Override.Match] picks first for a given
// runtime — silently dropping the assets of any sibling Override that differs
// only by variants (e.g. libc=musl vs libc=glibc).
func expandRuntimesByVariants(pkgInfo *registry.PackageInfo, rts []*runtime.Runtime) []*runtime.Runtime {
	if len(pkgInfo.Overrides) == 0 {
		return rts
	}
	keys := registry.SupportedVariantKeys()
	expanded := make([]*runtime.Runtime, 0, len(rts))
	seen := map[string]struct{}{}
	add := func(rt *runtime.Runtime) {
		k := runtimeKey(rt)
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
		for _, combo := range cartesianProduct(keys, valueSets) {
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

// collectVariantValueSets returns, for each supported variant key, the set of
// values to enumerate. An override that does not mention a key contributes the
// empty string (representing "no constraint" / fallback), so a fallback
// Override produces a runtime with that field cleared.
func collectVariantValueSets(overrides []*registry.Override, keys []string) map[string]map[string]struct{} {
	sets := make(map[string]map[string]struct{}, len(keys))
	for _, key := range keys {
		sets[key] = map[string]struct{}{}
	}
	for _, ov := range overrides {
		mentioned := make(map[string]string, len(ov.Variants))
		for _, v := range ov.Variants {
			mentioned[v.Key] = v.Value
		}
		for _, key := range keys {
			if val, ok := mentioned[key]; ok {
				sets[key][val] = struct{}{}
			} else {
				sets[key][""] = struct{}{}
			}
		}
	}
	return sets
}

func cartesianProduct(keys []string, sets map[string]map[string]struct{}) []map[string]string {
	combos := []map[string]string{{}}
	for _, key := range keys {
		values := sets[key]
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
