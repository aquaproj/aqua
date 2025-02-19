package genrgst

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/asset"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/ptr"
	"github.com/sirupsen/logrus"
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

func mergeGroups(pkg *registry.PackageInfo, groups []*Group) []string { //nolint:cyclop
	if len(groups) == 0 {
		return nil
	}
	versions := make([]string, 0, len(groups))
	for _, group := range groups {
		if len(group.releases) == 0 {
			continue
		}
		pkgInfo := group.pkg.Info
		vo := ConvertPkgToVO(pkgInfo)
		switch {
		case len(group.releases) == 1:
			v := group.releases[0].Tag
			vo.VersionConstraints = fmt.Sprintf(`Version == "%s"`, v)
			if !group.pkg.Info.NoAsset {
				versions = append(versions, v)
			}
		case group.fixed:
			tags := make([]string, len(group.releases))
			for i, release := range group.releases {
				tags[i] = fmt.Sprintf(`"%s"`, release.Tag)
			}
			vo.VersionConstraints = fmt.Sprintf("Version in [%s]", strings.Join(tags, ", "))
			if !group.pkg.Info.NoAsset {
				versions = append(versions, tags[0])
			}
		default:
			release := group.releases[len(group.releases)-1]
			vo.VersionConstraints = fmt.Sprintf(`semver("<= %s")`, strings.TrimPrefix(release.Version.String(), "v"))
			if !group.pkg.Info.NoAsset {
				versions = append(versions, release.Tag)
			}
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
	reverse(versions)
	return versions
}

func reverse(versions []string) {
	// sort.Reverse doesn't work well.
	// https://qiita.com/shibukawa/items/0e6e01dc41c352ccedb5
	size := len(versions)
	for i := range size / 2 {
		versions[i], versions[size-i-1] = versions[size-i-1], versions[i]
	}
}

func replaceVersion(assetName, version string) string {
	return strings.Replace(
		strings.Replace(assetName, version, "{{.Version}}", 1),
		strings.TrimPrefix(version, "v"), "{{trimV .Version}}", 1)
}

func groupByAllAsset(releases []*Release) []*Group {
	groups := []*Group{}
	var group *Group
	for _, release := range releases {
		assetNames := make([]string, len(release.assets))
		for i, asset := range release.assets {
			assetNames[i] = replaceVersion(asset.GetName(), release.Tag)
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
		return arr[i].releases[0].Tag < arr[j].releases[0].Tag
	})
	return arr
}

func sortAndMergeGroups(groups []*Group) []*Group {
	newGroups := make([]*Group, 0, len(groups))
	fixedGroups := make([]*Group, 0, len(groups))
	for _, group := range groups {
		if len(group.releases) == 1 {
			group.fixed = true
			fixedGroups = append(fixedGroups, group)
			continue
		}
		newGroups = append(newGroups, group)
	}
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
	if len(newGroups) == 0 {
		return []*Group{prevGroup}
	}
	if newGroups[len(newGroups)-1].allAsset != prevGroup.allAsset {
		return append(newGroups, prevGroup)
	}
	return newGroups
}

func (c *Controller) group(logE *logrus.Entry, pkgName string, releases []*Release) []*Group {
	if len(releases) == 0 {
		return nil
	}
	groups := groupByAllAsset(releases)
	excludeGroupsAssets(groups, pkgName)
	groups = groupByExcludedAsset(groups)

	for _, group := range groups {
		release := group.releases[0]
		pkgInfo := &registry.PackageInfo{}
		c.patchRelease(logE, pkgInfo, pkgName, release.Tag, release.assets)
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

	if newGroups[len(newGroups)-1].allAsset != prevGroup.allAsset {
		newGroups = append(newGroups, prevGroup)
	}
	group := newGroups[len(newGroups)-1]
	if group.pkg.Info.NoAsset || group.pkg.Info.Asset == "" {
		return newGroups[:len(newGroups)-1]
	}

	return sortAndMergeGroups(newGroups)
}

func (c *Controller) generatePackage(logE *logrus.Entry, pkgInfo *registry.PackageInfo, pkgName string, releases []*Release) []string {
	return mergeGroups(pkgInfo, c.group(logE, pkgName, releases))
}
