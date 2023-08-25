package asset

import (
	"reflect"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/ptr"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

func GetOSArch(goos, goarch string, assetInfos []*AssetInfo) *AssetInfo { //nolint:gocognit,cyclop
	var a, rawA *AssetInfo
	for _, assetInfo := range assetInfos {
		if (assetInfo.OS != goos || assetInfo.Arch != goarch) && !assetInfo.DarwinAll {
			continue
		}
		if assetInfo.Format == "" || assetInfo.Format == formatRaw { //nolint:nestif
			if rawA == nil {
				rawA = assetInfo
				continue
			}
			if assetInfo.Score > rawA.Score {
				rawA = assetInfo
				continue
			}
			rawAIdx := strings.Index(rawA.Template, "{")
			assetIdx := strings.Index(assetInfo.Template, "{")
			if rawAIdx != -1 && assetIdx != -1 {
				if rawAIdx > assetIdx {
					rawA = assetInfo
				}
				continue
			}
			if len(rawA.Template) > len(assetInfo.Template) {
				rawA = assetInfo
			}
			continue
		}
		if a == nil {
			a = assetInfo
			continue
		}
		if assetInfo.Score > a.Score {
			a = assetInfo
			continue
		}
		aIdx := strings.Index(a.Template, "{")
		assetIdx := strings.Index(assetInfo.Template, "{")
		if aIdx != -1 && assetIdx != -1 {
			if aIdx > assetIdx {
				a = assetInfo
			}
			continue
		}
		if len(a.Template) > len(assetInfo.Template) {
			a = assetInfo
		}
	}
	if a != nil {
		return a
	}
	return rawA
}

func mergeReplacements(goos string, m1, m2 map[string]string) (map[string]string, bool) {
	v1, ok1 := m1[goos]
	v2, ok2 := m2[goos]
	if (ok1 && ok2 && v1 == v2) || (!ok1 && !ok2) {
		m := map[string]string{}
		for k, v := range m1 {
			m[k] = v
		}
		for k, v := range m2 {
			m[k] = v
		}
		if len(m) == 0 {
			return nil, true
		}
		return m, true
	}
	return nil, false
}

func ParseAssetInfos(pkgInfo *registry.PackageInfo, assetInfos []*AssetInfo) { //nolint:funlen,gocognit,cyclop
	for _, goos := range []string{"linux", osDarwin, "windows"} {
		var overrides []*registry.Override
		var supportedEnvs []string
		for _, goarch := range []string{"amd64", "arm64"} {
			if assetInfo := GetOSArch(goos, goarch, assetInfos); assetInfo != nil {
				overrides = append(overrides, &registry.Override{
					GOOS:         assetInfo.OS,
					GOArch:       assetInfo.Arch,
					Format:       assetInfo.Format,
					Replacements: assetInfo.Replacements,
					Asset:        ptr.String(assetInfo.Template),
				})
				if goos == osDarwin && goarch == "amd64" {
					supportedEnvs = append(supportedEnvs, osDarwin)
				} else {
					supportedEnvs = append(supportedEnvs, goos+"/"+goarch)
				}
			}
		}
		if len(overrides) == 2 { //nolint:gomnd
			// amd64, arm64
			supportedEnvs = []string{goos}
			asset1 := overrides[0]
			asset2 := overrides[1]
			if asset1.Format == asset2.Format && *asset1.Asset == *asset2.Asset {
				// format and asset are equal
				replacements, ok := mergeReplacements(goos, overrides[0].Replacements, overrides[1].Replacements)
				if ok {
					overrides = []*registry.Override{
						{
							GOOS:         asset1.GOOS,
							Format:       asset1.Format,
							Replacements: replacements,
							Asset:        asset1.Asset,
						},
					}
				}
			}
		}
		if len(overrides) == 1 {
			overrides[0].GOArch = "" // omit arch
		}
		pkgInfo.Overrides = append(pkgInfo.Overrides, overrides...)
		pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, supportedEnvs...)
	}

	if checkRosetta2(assetInfos) {
		pkgInfo.Rosetta2 = ptr.Bool(true)
	}

	pkgInfo.SupportedEnvs = normalizeSupportedEnvs(pkgInfo.SupportedEnvs)

	pkgInfo.Format = getDefaultFormat(pkgInfo.Overrides)

	// Decide default Asset
	defaultAsset := getDefaultAsset(pkgInfo.Format, pkgInfo.Overrides)
	pkgInfo.Asset = &defaultAsset

	pkgInfo.Overrides = normalizeOverridesByAsset(*pkgInfo.Asset, pkgInfo.Overrides)

	pkgInfo.Replacements, pkgInfo.Overrides = normalizeOverridesByReplacements(pkgInfo)

	// Set CompleteWindowsExt
	for _, assetInfo := range assetInfos {
		if assetInfo.CompleteWindowsExt != nil {
			pkgInfo.CompleteWindowsExt = assetInfo.CompleteWindowsExt
			break
		}
	}
}

func normalizeOverridesByReplacements(pkgInfo *registry.PackageInfo) (map[string]string, []*registry.Override) { //nolint:funlen,gocognit,gocyclo,cyclop
	overrides := pkgInfo.Overrides
	var replacements map[string]string
	var ret []*registry.Override
	for _, override := range overrides {
		for k, v := range override.Replacements {
			score := -1
			matched := false
			isOS := runtime.IsOS(k)
			for _, override := range overrides {
				if v2, ok := override.Replacements[k]; ok { //nolint:nestif
					if v == v2 {
						score++
					} else if !matched {
						score--
					}
				} else {
					if override.GOOS == k || override.GOArch == k || (override.GOOS == "" && isOS) || (override.GOArch == "" && !isOS) {
						if !matched {
							score--
						}
					}
				}

				if isOS && override.GOOS == k && override.GOArch == "" {
					matched = true
				}
				if !isOS && override.GOArch == k && override.GOOS == "" {
					matched = true
				}
			}
			if score < 0 {
				continue
			}
			matched = false
			if replacements == nil {
				replacements = map[string]string{}
			}
			replacements[k] = v
			for _, override := range overrides {
				if v2, ok := override.Replacements[k]; ok { //nolint:nestif
					if v == v2 {
						delete(override.Replacements, k)
						if len(override.Replacements) == 0 {
							override.Replacements = nil
						}
					}
				} else {
					if override.GOOS == k || override.GOArch == k || (override.GOOS == "" && isOS) || (override.GOArch == "" && !isOS) {
						if !matched {
							if isKeyUsed(pkgInfo, override, k) {
								if override.Replacements == nil {
									override.Replacements = map[string]string{}
								}
								override.Replacements[k] = k
							}
						}
					}
				}

				if isOS && override.GOOS == k && override.GOArch == "" {
					matched = true
				}
				if !isOS && override.GOArch == k && override.GOOS == "" {
					matched = true
				}
			}
		}
		if override.Replacements != nil || override.Format != "" || override.Asset != nil {
			ret = append(ret, override)
		}
	}
	return replacements, ret
}

func isKeyUsed(pkgInfo *registry.PackageInfo, override *registry.Override, key string) bool {
	isOS := runtime.IsOS(key)
	var asset string
	if override.Asset != nil {
		asset = *override.Asset
	} else {
		asset = *pkgInfo.Asset
	}
	var checksumAsset string
	if override.Checksum != nil {
		checksumAsset = override.Checksum.Asset
	} else if pkgInfo.Checksum != nil {
		checksumAsset = pkgInfo.Checksum.Asset
	}
	if isOS {
		return strings.Contains(asset, "{{.OS}}") || strings.Contains(checksumAsset, "{{.OS}}")
	}
	return strings.Contains(asset, "{{.Arch}}") || strings.Contains(checksumAsset, "{{.Arch}}")
}

func normalizeOverridesByAsset(defaultAsset string, overrides []*registry.Override) []*registry.Override {
	ret := []*registry.Override{}
	for _, override := range overrides {
		if *override.Asset != defaultAsset {
			ret = append(ret, override)
			continue
		}
		override.Asset = nil
		if override.Format != "" || len(override.Replacements) != 0 {
			ret = append(ret, override)
		}
	}
	return ret
}

func getDefaultAsset(defaultFormat string, overrides []*registry.Override) string {
	assetCounts := map[string]int{}
	for _, override := range overrides {
		if override.Format != defaultFormat {
			continue
		}
		override.Format = ""
		assetCounts[*override.Asset]++
	}
	var maxCnt int
	var defaultAsset string
	for asset, cnt := range assetCounts {
		if cnt > maxCnt {
			defaultAsset = asset
			maxCnt = cnt
			continue
		}
	}
	return defaultAsset
}

func getDefaultFormat(overrides []*registry.Override) string {
	formatCounts := map[string]int{}
	for _, override := range overrides {
		formatCounts[override.Format]++
	}
	var maxCnt int
	var format string
	for f, cnt := range formatCounts {
		if (cnt > maxCnt) || (cnt == maxCnt && f != formatRaw) {
			format = f
			maxCnt = cnt
			continue
		}
	}
	return format
}

func checkRosetta2(assetInfos []*AssetInfo) bool {
	darwinAmd64 := GetOSArch(osDarwin, "amd64", assetInfos)
	darwinArm64 := GetOSArch(osDarwin, "arm64", assetInfos)
	return darwinAmd64 != nil && darwinArm64 == nil
}

func normalizeSupportedEnvs(envs registry.SupportedEnvs) []string {
	if reflect.DeepEqual(envs, registry.SupportedEnvs{"linux", osDarwin, "windows"}) {
		return nil
	}
	if reflect.DeepEqual(envs, registry.SupportedEnvs{"linux", osDarwin, "windows/amd64"}) {
		return []string{osDarwin, "linux", "amd64"}
	}
	if reflect.DeepEqual(envs, registry.SupportedEnvs{"linux/amd64", osDarwin, "windows/amd64"}) {
		return []string{osDarwin, "amd64"}
	}
	return envs
}

func ParseAssetName(assetName, version string) *AssetInfo { //nolint:cyclop
	assetInfo := &AssetInfo{
		Template: strings.Replace(assetName, version, "{{.Version}}", 1),
	}
	if assetInfo.Template == assetName {
		assetInfo.Template = strings.Replace(assetName, strings.TrimPrefix(version, "v"), "{{trimV .Version}}", 1)
	}
	lowAssetName := strings.ToLower(assetName)
	SetOS(assetName, lowAssetName, assetInfo)
	SetArch(assetName, lowAssetName, assetInfo)
	if assetInfo.Arch == "" && assetInfo.OS == osDarwin {
		if strings.Contains(lowAssetName, "_all") || strings.Contains(lowAssetName, "-all") || strings.Contains(lowAssetName, ".all") {
			assetInfo.DarwinAll = true
		}
		if strings.Contains(lowAssetName, "_universal") || strings.Contains(lowAssetName, "-universal") || strings.Contains(lowAssetName, ".universal") {
			assetInfo.DarwinAll = true
		}
	}
	assetInfo.Format = GetFormat(assetName)
	if assetInfo.Format != formatRaw {
		assetInfo.Template = assetInfo.Template[:len(assetInfo.Template)-len(assetInfo.Format)] + "{{.Format}}"
	}
	if assetInfo.OS == "windows" && assetInfo.Format == formatRaw {
		if strings.HasSuffix(assetInfo.Template, ".exe") {
			assetInfo.Template = strings.TrimSuffix(assetInfo.Template, ".exe")
		} else {
			assetInfo.CompleteWindowsExt = ptr.Bool(false)
		}
	}
	return assetInfo
}
