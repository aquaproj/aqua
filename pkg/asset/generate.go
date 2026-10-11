package asset

import (
	"maps"
	"reflect"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

// GetOSArch finds the best matching asset for a given OS and architecture combination.
// It scores assets based on format preferences and template complexity, preferring
// non-raw formats and simpler templates for better compatibility.
// GetOSArch returns the asset to install on one platform, or nil when the release
// has none for it.
//
// Several assets can claim the same platform, and they are ranked in this order: an
// archive beats a bare executable, a higher score beats a lower one, a template that
// becomes a template earlier beats one with a longer prefix, and a shorter template
// beats a longer one. The last rule is what tells the ordinary build from one for
// particular hardware, such as ollama-{{.OS}}-{{.Arch}}.{{.Format}} beside
// ollama-{{.OS}}-{{.Arch}}-rocm.{{.Format}}.
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
			if rawAIdx != assetIdx && rawAIdx != -1 && assetIdx != -1 {
				if rawAIdx > assetIdx {
					rawA = assetInfo
				}
				continue
			}
			if plainer(assetInfo.Template, rawA.Template) {
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
		if aIdx != assetIdx && aIdx != -1 && assetIdx != -1 {
			if aIdx > assetIdx {
				a = assetInfo
			}
			continue
		}
		if plainer(assetInfo.Template, a.Template) {
			a = assetInfo
		}
	}
	if a != nil {
		return a
	}
	return rawA
}

// plainer reports whether candidate is the ordinary build of what incumbent is a
// build of.
//
// A release can carry the same platform twice: ollama-{{.OS}}-{{.Arch}}.{{.Format}}
// beside ollama-{{.OS}}-{{.Arch}}-rocm.{{.Format}}, one for particular hardware. The
// one to install when nothing says otherwise is the one carrying nothing extra, and
// what makes it that is being the other with a piece taken off rather than merely
// being shorter.
//
// Being merely shorter is not enough. ollama's release also holds Ollama.dmg, the
// desktop application, whose template is shorter than the command's and is not the
// same name with less on it.
func plainer(candidate, incumbent string) bool {
	if len(candidate) >= len(incumbent) {
		return false
	}
	// Both end in the same format, so comparing the names means comparing what comes
	// before it.
	c, cFormat := splitFormat(candidate)
	i, iFormat := splitFormat(incumbent)
	if cFormat != iFormat {
		return false
	}
	return strings.HasPrefix(i, c)
}

// splitFormat separates a template's name from the format it ends with.
//
// The separator is the last dot outside a placeholder: the dots in {{.OS}} and
// {{.Format}} belong to the placeholder rather than to the name.
func splitFormat(template string) (string, string) {
	depth := 0
	for i := len(template) - 1; i >= 0; i-- {
		switch {
		case strings.HasPrefix(template[i:], "}}"):
			depth++
		case strings.HasPrefix(template[i:], "{{"):
			depth--
		case template[i] == '.' && depth == 0:
			return template[:i], template[i:]
		}
	}
	return template, ""
}

// mergeReplacements combines two replacement maps if they have compatible values for the given OS.
// Returns the merged map and true if merging is possible, otherwise nil and false.
func mergeReplacements(goos string, m1, m2 map[string]string) (map[string]string, bool) {
	v1, ok1 := m1[goos]
	v2, ok2 := m2[goos]
	if (ok1 && ok2 && v1 == v2) || (!ok1 && !ok2) {
		m := make(map[string]string, len(m1)+len(m2))
		maps.Copy(m, m1)
		maps.Copy(m, m2)
		if len(m) == 0 {
			return nil, true
		}
		return m, true
	}
	return nil, false
}

// ParseAssetInfos processes asset information and populates package registry metadata.
// It generates overrides for different OS/architecture combinations, determines supported
// environments, and optimizes the configuration for cross-platform compatibility.
func ParseAssetInfos(pkgInfo *registry.PackageInfo, assetInfos []*AssetInfo) { //nolint:funlen,cyclop
	for _, goos := range []string{osLinux, osDarwin, osWindows} {
		var overrides []*registry.Override
		var supportedEnvs []string
		for _, goarch := range []string{archAmd64, archArm64} {
			if assetInfo := GetOSArch(goos, goarch, assetInfos); assetInfo != nil {
				overrides = append(overrides, &registry.Override{
					GOOS:         assetInfo.OS,
					GOArch:       assetInfo.Arch,
					Format:       assetInfo.Format,
					Replacements: assetInfo.Replacements,
					Asset:        assetInfo.Template,
				})
				if goos == osDarwin && goarch == archAmd64 {
					supportedEnvs = append(supportedEnvs, osDarwin)
				} else {
					supportedEnvs = append(supportedEnvs, goos+"/"+goarch)
				}
				if assetInfo.CompleteWindowsExt != nil {
					pkgInfo.CompleteWindowsExt = assetInfo.CompleteWindowsExt
				}
			}
		}
		if len(overrides) == 2 { //nolint:mnd
			// amd64, arm64
			supportedEnvs = []string{goos}
			asset1 := overrides[0]
			asset2 := overrides[1]
			if asset1.Format == asset2.Format && asset1.Asset == asset2.Asset {
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
		pkgInfo.Rosetta2 = true
	}
	if checkWindowsARMEmulation(assetInfos) {
		pkgInfo.WindowsARMEmulation = true
	}

	pkgInfo.SupportedEnvs = normalizeSupportedEnvs(pkgInfo.SupportedEnvs)

	pkgInfo.Format = getDefaultFormat(pkgInfo.Overrides)

	// Decide default Asset
	defaultAsset := getDefaultAsset(pkgInfo.Format, pkgInfo.Overrides)
	pkgInfo.Asset = defaultAsset

	clearDefaultAsset(pkgInfo.Asset, pkgInfo.Overrides)

	pkgInfo.Replacements, pkgInfo.Overrides = normalizeOverridesByReplacements(pkgInfo)

	// After the replacements are settled, because a musl override is the first that
	// matches on a musl system and has to say whatever the package-level replacements
	// would get wrong for its asset.
	pkgInfo.Overrides = append(muslVariants(pkgInfo, assetInfos), pkgInfo.Overrides...)
}

// muslVariants returns overrides choosing the musl build on a musl system, for a release
// that publishes one beside the Linux build chosen for the package.
//
// Which libc a build needs is in its name only as "musl", so this is what lets aqua gr
// write the variant a person had to: UsefulSoftwareCo/executor publishes
// executor-linux-x64.tar.gz and executor-linux-x64-musl.tar.gz. A release whose Linux
// build is already the musl one, as a Rust release with both target triples is, needs
// nothing: a static build runs on glibc too.
//
// The overrides go first, because aqua applies the first override that matches, and so
// say everything their asset needs rather than relying on the Linux override. Before the
// musl ones go glibc ones saying what the Linux override does: aqua < v2.58.0 doesn't
// know variants, matches the first override for linux whatever its variant, and has to
// stay on the build it always installed.
func muslVariants(pkgInfo *registry.PackageInfo, assetInfos []*AssetInfo) []*registry.Override {
	builds := muslBuilds(assetInfos)
	if len(builds) == 0 {
		return nil
	}
	amd, arm := builds[archAmd64], builds[archArm64]
	scopes := [][]string{}
	if amd != nil && arm != nil && amd.Template == arm.Template && amd.Format == arm.Format &&
		spelling(amd.Replacements, osLinux) == spelling(arm.Replacements, osLinux) {
		scopes = append(scopes, []string{archAmd64, archArm64})
	} else {
		for _, goarch := range []string{archAmd64, archArm64} {
			if builds[goarch] != nil {
				scopes = append(scopes, []string{goarch})
			}
		}
	}
	var glibc, musl []*registry.Override
	for _, goarchs := range scopes {
		glibc = append(glibc, glibcOverrides(pkgInfo.Overrides, goarchs)...)
		musl = append(musl, muslOverride(pkgInfo, builds, goarchs))
	}
	return append(glibc, musl...)
}

// glibcOverrides returns, for the architectures a musl override covers, overrides choosing
// on glibc what the Linux override chose before there was a musl one.
//
// They come before the musl override, so that an aqua that doesn't know variants matches
// them first. Architectures that the Linux override treats differently get one each.
func glibcOverrides(overrides []*registry.Override, goarchs []string) []*registry.Override {
	variants := []*registry.Variant{{Key: "libc", Value: "glibc"}}
	first := linuxOverride(overrides, goarchs[0])
	shared := true
	for _, goarch := range goarchs[1:] {
		if linuxOverride(overrides, goarch) != first {
			shared = false
		}
	}
	ret := []*registry.Override{}
	for i, goarch := range goarchs {
		if shared && i > 0 {
			break
		}
		ov := &registry.Override{}
		if base := linuxOverride(overrides, goarch); base != nil {
			*ov = *base
		}
		ov.GOOS = osLinux
		ov.GOArch = ""
		if len(goarchs) == 1 || !shared {
			ov.GOArch = goarch
		}
		ov.Variants = variants
		ret = append(ret, ov)
	}
	return ret
}

// linuxOverride is the override aqua applies on Linux for goarch, or nil when none does.
func linuxOverride(overrides []*registry.Override, goarch string) *registry.Override {
	for _, ov := range overrides {
		if len(ov.Variants) != 0 {
			continue
		}
		if (ov.GOOS == "" || ov.GOOS == osLinux) && (ov.GOArch == "" || ov.GOArch == goarch) {
			return ov
		}
	}
	return nil
}

// muslBuilds returns, by architecture, the musl build of each Linux architecture whose
// chosen build isn't one.
func muslBuilds(assetInfos []*AssetInfo) map[string]*AssetInfo {
	musl := make([]*AssetInfo, 0, len(assetInfos))
	for _, a := range assetInfos {
		if a.Musl {
			musl = append(musl, a)
		}
	}
	if len(musl) == 0 {
		return nil
	}
	builds := map[string]*AssetInfo{}
	for _, goarch := range []string{archAmd64, archArm64} {
		chosen := GetOSArch(osLinux, goarch, assetInfos)
		if chosen == nil || chosen.Musl {
			continue
		}
		if m := GetOSArch(osLinux, goarch, musl); m != nil {
			builds[goarch] = m
		}
	}
	return builds
}

// muslOverride is the override naming the musl build of the given architectures, each of
// which spells its architecture as its own build does.
//
// It says a replacement only where the package's would spell the OS or an architecture
// differently from the build's name, and a format only where it differs from the
// package's.
func muslOverride(pkgInfo *registry.PackageInfo, builds map[string]*AssetInfo, goarchs []string) *registry.Override {
	build := builds[goarchs[0]]
	ov := &registry.Override{
		GOOS:     osLinux,
		Asset:    build.Template,
		Variants: []*registry.Variant{{Key: "libc", Value: "musl"}},
	}
	if len(goarchs) == 1 {
		ov.GOArch = goarchs[0]
	}
	if build.Format != pkgInfo.Format {
		ov.Format = build.Format
	}
	want := map[string]string{}
	if strings.Contains(build.Template, "{{.OS}}") {
		want[osLinux] = spelling(build.Replacements, osLinux)
	}
	if strings.Contains(build.Template, "{{.Arch}}") {
		for _, goarch := range goarchs {
			want[goarch] = spelling(builds[goarch].Replacements, goarch)
		}
	}
	for k, v := range want {
		if spelling(pkgInfo.Replacements, k) == v {
			continue
		}
		if ov.Replacements == nil {
			ov.Replacements = map[string]string{}
		}
		ov.Replacements[k] = v
	}
	return ov
}

// spelling is how replacements spell key, which is key itself when they don't replace it.
func spelling(replacements map[string]string, key string) string {
	if v, ok := replacements[key]; ok {
		return v
	}
	return key
}

// normalizeOverridesByReplacements extracts common replacements from overrides to reduce duplication.
// It analyzes replacement patterns across overrides and promotes frequently used replacements
// to the package level, simplifying the configuration.
func normalizeOverridesByReplacements(pkgInfo *registry.PackageInfo) (map[string]string, []*registry.Override) { //nolint:funlen,gocognit,gocyclo,cyclop
	overrides := pkgInfo.Overrides
	var replacements map[string]string
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
								if k != archArm64 || (override.GOOS != osWindows || !pkgInfo.WindowsARMEmulation) && (override.GOOS != osDarwin || !pkgInfo.Rosetta2) {
									// https://github.com/aquaproj/aqua/issues/3336
									// The map is made only when something goes in it, so an
									// override left with nothing is dropped below.
									if override.Replacements == nil {
										override.Replacements = map[string]string{}
									}
									override.Replacements[k] = k
								}
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
	}
	// Kept after every replacement is promoted, not as each override is looked at:
	// promoting one can give an override looked at earlier a replacement of its own,
	// which cancels the promoted one for it. zk-org/zk names linux amd64 but darwin and
	// windows x86_64, and a linux override dropped before amd64: x86_64 was promoted
	// left linux with x86_64 too.
	var ret []*registry.Override
	for _, override := range overrides {
		if len(override.Replacements) != 0 || override.Format != "" || override.Asset != "" {
			ret = append(ret, override)
		}
	}
	return replacements, ret
}

// isKeyUsed determines if a template key (OS or Arch) is referenced in asset templates.
// It checks both the main asset template and checksum asset template for template variables.
func isKeyUsed(pkgInfo *registry.PackageInfo, override *registry.Override, key string) bool {
	isOS := runtime.IsOS(key)
	var asset string
	if override.Asset != "" {
		asset = override.Asset
	} else {
		asset = pkgInfo.Asset
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

// clearDefaultAsset clears the asset of the overrides that use the default one.
//
// An override left saying nothing is kept: normalizeOverridesByReplacements may give it a
// replacement of its own when it promotes one to the package, and drops it if it doesn't.
func clearDefaultAsset(defaultAsset string, overrides []*registry.Override) {
	for _, override := range overrides {
		if override.Asset == defaultAsset {
			override.Asset = ""
		}
	}
}

// getDefaultAsset determines the most commonly used asset template for the default format.
// It counts asset usage and returns the most frequent one as the default.
func getDefaultAsset(defaultFormat string, overrides []*registry.Override) string {
	assetCounts := map[string]int{}
	for _, override := range overrides {
		if override.Format != defaultFormat {
			continue
		}
		override.Format = ""
		assetCounts[override.Asset]++
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

// getDefaultFormat determines the most commonly used format across all overrides.
// It prefers non-raw formats when counts are equal to optimize for archive handling.
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

// checkRosetta2 determines if a package supports Rosetta2 emulation on macOS.
// Returns true if only amd64 assets are available for Darwin with arch templating.
func checkRosetta2(assetInfos []*AssetInfo) bool {
	darwinAmd64 := GetOSArch(osDarwin, archAmd64, assetInfos)
	darwinArm64 := GetOSArch(osDarwin, archArm64, assetInfos)
	return darwinAmd64 != nil && darwinArm64 == nil && strings.Contains(darwinAmd64.Template, "{{.Arch}}")
}

// checkWindowsARMEmulation determines if a package supports ARM emulation on Windows.
// Returns true if only amd64 assets are available for Windows with arch templating.
func checkWindowsARMEmulation(assetInfos []*AssetInfo) bool {
	windowsAmd64 := GetOSArch(osWindows, archAmd64, assetInfos)
	windowsArm64 := GetOSArch(osWindows, archArm64, assetInfos)
	return windowsAmd64 != nil && windowsArm64 == nil && strings.Contains(windowsAmd64.Template, "{{.Arch}}")
}

// normalizeSupportedEnvs simplifies supported environment lists by removing redundant entries.
// It collapses common platform combinations into more concise representations.
func normalizeSupportedEnvs(envs registry.SupportedEnvs) []string {
	if reflect.DeepEqual(envs, registry.SupportedEnvs{osLinux, osDarwin, osWindows}) {
		return nil
	}
	if reflect.DeepEqual(envs, registry.SupportedEnvs{osLinux, osDarwin, "windows/amd64"}) {
		return nil
	}
	if reflect.DeepEqual(envs, registry.SupportedEnvs{"linux/amd64", osDarwin, "windows/amd64"}) {
		return []string{osDarwin, osWindows, archAmd64}
	}
	return envs
}

// ParseAssetName analyzes a release asset name to extract platform, architecture, and format information.
// It generates templates with placeholders for version, OS, architecture, and format,
// enabling dynamic asset URL generation for different platforms.
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
	if assetInfo.OS == osLinux && strings.Contains(lowAssetName, "musl") {
		assetInfo.Musl = true
	}
	assetInfo.Format = getFormat(assetName)
	if assetInfo.Format != formatRaw {
		assetInfo.Template = assetInfo.Template[:len(assetInfo.Template)-len(assetInfo.Format)] + "{{.Format}}"
	}
	if assetInfo.OS == osWindows && assetInfo.Format == formatRaw {
		if before, ok := strings.CutSuffix(assetInfo.Template, ".exe"); ok {
			assetInfo.Template = before
		} else {
			assetInfo.CompleteWindowsExt = new(false)
		}
	}
	return assetInfo
}
