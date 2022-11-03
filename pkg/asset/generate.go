package asset

import (
	"reflect"
	"strings"

	"github.com/aquaproj/aqua/pkg/config/registry"
)

func boolP(b bool) *bool {
	return &b
}

func strP(s string) *string {
	return &s
}

func GetOSArch(goos, goarch string, assetInfos []*AssetInfo) *AssetInfo { //nolint:gocognit,cyclop
	var a, rawA *AssetInfo
	for _, assetInfo := range assetInfos {
		assetInfo := assetInfo
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

func mergeReplacements(m1, m2 map[string]string) (map[string]string, bool) {
	if len(m1) == 0 {
		return m2, true
	}
	if len(m2) == 0 {
		return m1, true
	}
	m := map[string]string{}
	for k, v1 := range m1 {
		m[k] = v1
		if v2, ok := m2[k]; ok && v1 != v2 {
			return nil, false
		}
	}
	for k, v2 := range m2 {
		if _, ok := m[k]; !ok {
			m[k] = v2
		}
	}
	return m, true
}

func ParseAssetInfos(pkgInfo *registry.PackageInfo, assetInfos []*AssetInfo) { //nolint:funlen,gocognit,cyclop,gocyclo
	for _, goos := range []string{"linux", "darwin", "windows"} {
		var overrides []*registry.Override
		var supportedEnvs []string
		for _, goarch := range []string{"amd64", "arm64"} {
			if assetInfo := GetOSArch(goos, goarch, assetInfos); assetInfo != nil {
				overrides = append(overrides, &registry.Override{
					GOOS:         assetInfo.OS,
					GOArch:       assetInfo.Arch,
					Format:       assetInfo.Format,
					Replacements: assetInfo.Replacements,
					Asset:        strP(assetInfo.Template),
				})
				if goos == "darwin" && goarch == "amd64" {
					supportedEnvs = append(supportedEnvs, "darwin")
				} else {
					supportedEnvs = append(supportedEnvs, goos+"/"+goarch)
				}
			}
		}
		if len(overrides) == 2 { //nolint:gomnd
			supportedEnvs = []string{goos}
			asset1 := overrides[0]
			asset2 := overrides[1]
			if asset1.Format == asset2.Format && *asset1.Asset == *asset2.Asset {
				replacements, ok := mergeReplacements(overrides[0].Replacements, overrides[1].Replacements)
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
			overrides[0].GOArch = ""
		}
		pkgInfo.Overrides = append(pkgInfo.Overrides, overrides...)
		pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, supportedEnvs...)
	}

	darwinAmd64 := GetOSArch("darwin", "amd64", assetInfos)
	darwinArm64 := GetOSArch("darwin", "arm64", assetInfos)
	if darwinAmd64 != nil && darwinArm64 == nil {
		pkgInfo.Rosetta2 = boolP(true)
	}

	if reflect.DeepEqual(pkgInfo.SupportedEnvs, registry.SupportedEnvs{"linux", "darwin", "windows"}) {
		pkgInfo.SupportedEnvs = nil
	}
	if reflect.DeepEqual(pkgInfo.SupportedEnvs, registry.SupportedEnvs{"linux", "darwin", "windows/amd64"}) {
		pkgInfo.SupportedEnvs = registry.SupportedEnvs{"darwin", "linux", "amd64"}
	}
	if reflect.DeepEqual(pkgInfo.SupportedEnvs, registry.SupportedEnvs{"linux/amd64", "darwin", "windows/amd64"}) {
		pkgInfo.SupportedEnvs = registry.SupportedEnvs{"darwin", "amd64"}
	}

	formatCounts := map[string]int{}
	for _, override := range pkgInfo.Overrides {
		formatCounts[override.Format]++
	}
	maxCnt := 0
	for f, cnt := range formatCounts {
		if cnt > maxCnt {
			pkgInfo.Format = f
			maxCnt = cnt
			continue
		}
		if cnt == maxCnt && f != formatRaw {
			pkgInfo.Format = f
			maxCnt = cnt
			continue
		}
	}
	assetCounts := map[string]int{}
	for _, override := range pkgInfo.Overrides {
		override := override
		if override.Format != pkgInfo.Format {
			continue
		}
		override.Format = ""
		assetCounts[*override.Asset]++
	}
	maxCnt = 0
	for asset, cnt := range assetCounts {
		asset := asset
		if cnt > maxCnt {
			pkgInfo.Asset = &asset
			maxCnt = cnt
			continue
		}
	}
	overrides := []*registry.Override{}
	for _, override := range pkgInfo.Overrides {
		override := override
		if *override.Asset != *pkgInfo.Asset {
			overrides = append(overrides, override)
			continue
		}
		override.Asset = nil
		if override.Format != "" || len(override.Replacements) != 0 {
			overrides = append(overrides, override)
		}
	}
	pkgInfo.Overrides = overrides

	overrides = []*registry.Override{}
	for _, override := range pkgInfo.Overrides {
		override := override
		if len(override.Replacements) == 0 {
			overrides = append(overrides, override)
			continue
		}
		if pkgInfo.Replacements == nil {
			pkgInfo.Replacements = registry.Replacements{}
		}
		for k, v := range override.Replacements {
			vp, ok := pkgInfo.Replacements[k]
			if !ok {
				pkgInfo.Replacements[k] = v
				delete(override.Replacements, k)
				continue
			}
			if v == vp {
				delete(override.Replacements, k)
				continue
			}
		}
		if len(override.Replacements) != 0 || override.Format != "" || override.Asset != nil {
			overrides = append(overrides, override)
		}
	}
	pkgInfo.Overrides = overrides
	if len(pkgInfo.Overrides) == 0 && pkgInfo.Format != "" && pkgInfo.Format != formatRaw {
		asset := strings.Replace(*pkgInfo.Asset, "{{.Format}}", pkgInfo.Format, 1)
		pkgInfo.Asset = &asset
		pkgInfo.Format = ""
	}
	for _, assetInfo := range assetInfos {
		if assetInfo.CompleteWindowsExt != nil {
			pkgInfo.CompleteWindowsExt = assetInfo.CompleteWindowsExt
			break
		}
	}
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
	if assetInfo.Arch == "" && assetInfo.OS == "darwin" {
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
			assetInfo.CompleteWindowsExt = boolP(false)
		}
	}
	return assetInfo
}
