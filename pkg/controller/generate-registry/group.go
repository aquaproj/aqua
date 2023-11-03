package genrgst

import (
	"fmt"
	"reflect"
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
}

func mergeGroups(groups []*Group) (*registry.PackageInfo, []string) {
	pkg := &registry.PackageInfo{}
	versions := make([]string, 0, len(groups))
	for _, group := range groups {
		if len(group.releases) == 0 {
			continue
		}
		pkgInfo := group.pkg.Info
		vo := &registry.VersionOverride{
			Asset:              pkgInfo.Asset,
			Files:              pkgInfo.Files,
			Format:             pkgInfo.Format,
			Overrides:          pkgInfo.Overrides,
			Replacements:       pkgInfo.Replacements,
			SupportedEnvs:      pkgInfo.SupportedEnvs,
			CompleteWindowsExt: pkgInfo.CompleteWindowsExt,
			Checksum:           pkgInfo.Checksum,
		}
		if len(group.releases) == 1 {
			v := group.releases[0].Tag
			vo.VersionConstraints = fmt.Sprintf(`Version == "%s"`, v)
			if !group.pkg.Info.NoAsset {
				versions = append(versions, v)
			}
		} else {
			v := group.releases[len(group.releases)-1].Tag
			vo.VersionConstraints = fmt.Sprintf(`semver("<= %s")`, v)
			if !group.pkg.Info.NoAsset {
				versions = append(versions, v)
			}
		}
		if pkgInfo.NoAsset {
			vo.NoAsset = ptr.Bool(true)
		}
		if pkgInfo.Rosetta2 {
			vo.Rosetta2 = ptr.Bool(true)
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
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	return pkg, versions
}

func replaceVersion(assetName, version string) string {
	return strings.Replace(
		strings.Replace(assetName, version, "{{.Version}}", 1),
		strings.TrimPrefix(version, "v"), "{{trimV .Version}}", 1)
}

func (c *Controller) group(logE *logrus.Entry, pkgName string, releases []*Release) []*Group { //nolint:funlen,cyclop
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
	if groups[len(groups)-1].allAsset != group.allAsset {
		groups = append(groups, group)
	}

	var prevGroup *Group
	newGroups := make([]*Group, 0, len(groups))
	for _, group := range groups {
		assetNames := make([]string, 0, len(group.assetNames))
		for _, assetName := range group.assetNames {
			if asset.Exclude(pkgName, assetName) {
				continue
			}
			assetNames = append(assetNames, assetName)
		}
		group.assetNames = assetNames
		group.allAsset = strings.Join(assetNames, "\n")
		if prevGroup == nil {
			prevGroup = group
			continue
		}
		if prevGroup.allAsset == group.allAsset {
			prevGroup.releases = append(prevGroup.releases, group.releases...)
			continue
		}
		newGroups = append(newGroups, prevGroup)
		prevGroup = group
	}
	if newGroups[len(newGroups)-1].allAsset != prevGroup.allAsset {
		newGroups = append(newGroups, prevGroup)
	}

	for _, group := range newGroups {
		release := group.releases[0]
		pkgInfo := &registry.PackageInfo{}
		c.patchRelease(logE, pkgInfo, pkgName, release.Tag, release.assets)
		group.pkg = &Package{
			Info:    pkgInfo,
			Version: release.Tag,
			SemVer:  release.Tag,
		}
	}

	if len(newGroups) == 1 {
		return newGroups
	}
	prevGroup = newGroups[0]
	groups = newGroups
	newGroups = []*Group{prevGroup}
	for _, group := range newGroups[1:] {
		if reflect.DeepEqual(group.pkg, prevGroup.pkg) {
			prevGroup.releases = append(prevGroup.releases, group.releases...)
			continue
		}
		groups = append(groups, group)
		prevGroup = group
	}

	if groups[len(groups)-1].allAsset != prevGroup.allAsset {
		groups = append(groups, prevGroup)
	}
	group = groups[len(groups)-1]
	if group.pkg.Info.NoAsset || group.pkg.Info.Asset == "" {
		return groups[:len(groups)-1]
	}

	return groups
}

func (c *Controller) generatePackage(logE *logrus.Entry, pkgName string, releases []*Release) (*registry.PackageInfo, []string) {
	groups := c.group(logE, pkgName, releases)
	pkgInfo, versions := mergeGroups(groups)
	return pkgInfo, versions
}
