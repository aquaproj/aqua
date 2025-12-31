package genrgst

import (
	"fmt"
	"log/slog"
	"maps"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/asset"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/ptr"
)

type Group struct {
	releases   []*Release
	allAsset   string
	pkg        *Package
	assetNames []string
	fixed      bool
}

func ConvertPkgToVO(pkgInfo *registry.PackageInfo) *registry.VersionOverride {
	return &registry.VersionOverride{
		Asset:              pkgInfo.Asset,
		Files:              pkgInfo.Files,
		Format:             pkgInfo.Format,
		Overrides:          pkgInfo.Overrides,
		Replacements:       pkgInfo.Replacements,
		SupportedEnvs:      pkgInfo.SupportedEnvs,
		CompleteWindowsExt: pkgInfo.CompleteWindowsExt,
		Checksum:           pkgInfo.Checksum,
		SLSAProvenance:     pkgInfo.SLSAProvenance,
	}
}

func toVersionInString(releases []*Release) string {
	tags := make([]string, len(releases))
	for i, release := range releases {
		tags[i] = fmt.Sprintf(`"%s"`, release.Tag)
	}
	return fmt.Sprintf("Version in [%s]", strings.Join(tags, ", "))
}

func (g *Group) VersionConstraint() (string, *Release) { //nolint:cyclop
	switch {
	case len(g.releases) == 1:
		v := g.releases[0].Tag
		vc := fmt.Sprintf(`Version == "%s"`, v)
		if !g.pkg.Info.NoAsset {
			return vc, g.releases[0]
		}
		return vc, nil
	case g.fixed:
		vc := toVersionInString(g.releases)
		if !g.pkg.Info.NoAsset {
			return vc, g.releases[0]
		}
		return vc, nil
	default:
		nonSemvers := make([]*Release, 0, len(g.releases))
		for i := len(g.releases) - 1; i >= 0; i-- {
			release := g.releases[len(g.releases)-1]
			if release.Version == nil {
				nonSemvers = append(nonSemvers, release)
				continue
			}
			var vc string
			if len(nonSemvers) > 0 {
				vc = fmt.Sprintf(`semver("<= %s") or %s`, strings.TrimPrefix(release.Version.String(), "v"), toVersionInString(nonSemvers))
			} else {
				vc = fmt.Sprintf(`semver("<= %s")`, strings.TrimPrefix(release.Version.String(), "v"))
			}
			if !g.pkg.Info.NoAsset {
				return vc, release
			}
			return vc, nil
		}
		vc := toVersionInString(nonSemvers)
		if g.pkg.Info.NoAsset {
			return vc, nil
		}
		return vc, g.releases[0]
	}
}

func mergeGroups(pkg *registry.PackageInfo, groups []*Group) []string { //nolint:cyclop
	if len(groups) == 0 {
		return nil
	}
	releases := make([]*Release, 0, len(groups))
	for _, group := range groups {
		if len(group.releases) == 0 {
			continue
		}
		pkgInfo := group.pkg.Info
		vo := ConvertPkgToVO(pkgInfo)
		var release *Release
		vo.VersionConstraints, release = group.VersionConstraint()
		if release != nil {
			releases = append(releases, release)
		}
		if pkgInfo.NoAsset {
			vo.NoAsset = ptr.Bool(true)
		}
		if pkgInfo.Rosetta2 {
			vo.Rosetta2 = ptr.Bool(true)
		}
		if pkgInfo.WindowsARMEmulation {
			vo.WindowsARMEmulation = ptr.Bool(true)
		}
		if pkgInfo.VersionFilter != "" {
			vo.VersionFilter = ptr.String(pkgInfo.VersionFilter)
		}
		if pkgInfo.VersionPrefix != "" {
			vo.VersionPrefix = ptr.String(pkgInfo.VersionPrefix)
		}
		pkg.VersionOverrides = append(pkg.VersionOverrides, vo)
	}
	pkg.VersionOverrides[len(pkg.VersionOverrides)-1].VersionConstraints = "true"
	sort.Slice(releases, func(i, j int) bool {
		rI := releases[i]
		rJ := releases[j]
		return rI.GreaterThan(rJ)
	})
	versions := make([]string, len(releases))
	for i, release := range releases {
		versions[i] = release.Tag
	}
	return versions
}

func replaceVersion(assetName, version, semver string) string {
	s := strings.ReplaceAll(
		strings.ReplaceAll(assetName, version, "{{.Version}}"),
		strings.TrimPrefix(version, "v"), "{{trimV .Version}}")
	if semver == version {
		return s
	}
	return strings.ReplaceAll(s, semver, "{{.SemVer}}")
}

func groupByAllAsset(releases []*Release) []*Group {
	groups := []*Group{}
	var group *Group
	for _, release := range releases {
		assetNames := make([]string, len(release.assets))
		semver := strings.TrimPrefix(release.Tag, release.VersionPrefix)
		for i, asset := range release.assets {
			assetNames[i] = replaceVersion(asset.GetName(), release.Tag, semver)
		}
		sort.Strings(assetNames)
		allAsset := strings.Join(assetNames, "\n")
		if group == nil {
			group = &Group{
				releases: []*Release{
					release,
				},
				allAsset:   allAsset,
				assetNames: assetNames,
			}
			continue
		}
		if group.allAsset == allAsset {
			group.releases = append(group.releases, release)
			continue
		}
		groups = append(groups, group)
		group = &Group{
			releases: []*Release{
				release,
			},
			allAsset:   allAsset,
			assetNames: assetNames,
		}
	}
	if len(groups) == 0 && group != nil {
		groups = append(groups, group)
	}
	if groups[len(groups)-1].allAsset != group.allAsset {
		groups = append(groups, group)
	}
	return groups
}

func mergeFixedGroups(groups []*Group) []*Group {
	m := map[string]*Group{}
	for _, group := range groups {
		a, ok := m[group.allAsset]
		if !ok {
			m[group.allAsset] = group
			continue
		}
		a.releases = append(a.releases, group.releases...)
	}
	arr := slices.Collect(maps.Values(m))
	sort.Slice(arr, func(i, j int) bool {
		rI := arr[i].releases[0]
		rJ := arr[j].releases[0]
		return rI.LessThan(rJ)
	})

	// Move the group with NoAsset to the top.
	for i, a := range arr {
		if a.pkg.Info.NoAsset {
			return append(append([]*Group{a}, arr[:i]...), arr[i+1:]...)
		}
	}
	return arr
}

func sortAndMergeGroups(groups []*Group) []*Group {
	newGroups := make([]*Group, 0, len(groups))
	fixedGroups := make([]*Group, 0, len(groups))
	lastIdx := len(groups) - 1
	if lastIdx < 0 {
		return groups
	}
	for _, group := range groups[:lastIdx] {
		if len(group.releases) == 1 {
			group.fixed = true
			fixedGroups = append(fixedGroups, group)
			continue
		}
		newGroups = append(newGroups, group)
	}
	newGroups = append(newGroups, groups[lastIdx])
	return append(mergeFixedGroups(fixedGroups), groupByExcludedAsset(newGroups)...)
}

func excludeGroupAssets(group *Group, pkgName string) {
	assetNames := make([]string, 0, len(group.assetNames))
	for _, assetName := range group.assetNames {
		if asset.Exclude(pkgName, assetName) {
			continue
		}
		assetNames = append(assetNames, assetName)
	}
	group.assetNames = assetNames
	group.allAsset = strings.Join(assetNames, "\n")
}

func excludeGroupsAssets(groups []*Group, pkgName string) {
	for _, group := range groups {
		excludeGroupAssets(group, pkgName)
	}
}

func groupByExcludedAsset(groups []*Group) []*Group {
	if len(groups) == 0 {
		return groups
	}
	newGroups := make([]*Group, 0, len(groups))
	prevGroup := groups[0]
	for _, group := range groups[1:] {
		if prevGroup.allAsset == group.allAsset {
			prevGroup.releases = append(prevGroup.releases, group.releases...)
			continue
		}
		if prevGroup.pkg != nil && group.pkg != nil && reflect.DeepEqual(prevGroup.pkg.Info, group.pkg.Info) {
			prevGroup.releases = append(prevGroup.releases, group.releases...)
			continue
		}
		newGroups = append(newGroups, prevGroup)
		prevGroup = group
	}
	if len(newGroups) == 0 || newGroups[len(newGroups)-1].allAsset != prevGroup.allAsset {
		return append(newGroups, prevGroup)
	}
	return newGroups
}

func (c *Controller) group(logger *slog.Logger, pkgInfo *registry.PackageInfo, pkgName string, releases []*Release) []*Group {
	if len(releases) == 0 {
		return nil
	}
	groups := groupByAllAsset(releases)
	excludeGroupsAssets(groups, pkgName)
	groups = groupByExcludedAsset(groups)

	for _, group := range groups {
		release := group.releases[0]
		pkgInfo := &registry.PackageInfo{
			RepoOwner: pkgInfo.RepoOwner,
			RepoName:  pkgInfo.RepoName,
		}
		c.patchRelease(logger, pkgInfo, pkgName, release.Tag, group.assetNames)
		group.pkg = &Package{
			Info:    pkgInfo,
			Version: release.Tag,
			SemVer:  release.Tag,
		}
	}

	if len(groups) == 1 {
		return groups
	}
	prevGroup := groups[0]
	newGroups := make([]*Group, 0, len(groups))
	for _, group := range groups[1:] {
		if reflect.DeepEqual(group.pkg.Info, prevGroup.pkg.Info) {
			prevGroup.releases = append(prevGroup.releases, group.releases...)
			continue
		}
		newGroups = append(newGroups, prevGroup)
		prevGroup = group
	}

	if len(newGroups) == 0 || newGroups[len(newGroups)-1].allAsset != prevGroup.allAsset {
		newGroups = append(newGroups, prevGroup)
	}
	group := newGroups[len(newGroups)-1]
	if group.pkg.Info.NoAsset || group.pkg.Info.Asset == "" {
		return newGroups[:len(newGroups)-1]
	}

	return sortAndMergeGroups(newGroups)
}

func (c *Controller) generatePackage(logger *slog.Logger, pkgInfo *registry.PackageInfo, pkgName string, releases []*Release) []string {
	return mergeGroups(pkgInfo, c.group(logger, pkgInfo, pkgName, releases))
}
