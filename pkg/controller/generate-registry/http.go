package genrgst

import (
	"sort"

	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func convHTTPReleases(logE *logrus.Entry, cfg *Config, hrs map[string][]string) []*Release {
	releases := make([]*Release, 0, len(hrs))
	for tag := range hrs {
		if excludeVersion(logE, tag, cfg) {
			continue
		}
		v, prefix, err := versiongetter.GetVersionAndPrefix(tag)
		if err != nil {
			logerr.WithError(logE, err).WithField("tag_name", tag).Warn("parse a tag as semver")
		}
		releases = append(releases, &Release{
			Tag:           tag,
			Version:       v,
			VersionPrefix: prefix,
		})
	}
	sort.Slice(releases, func(i, j int) bool {
		r1 := releases[i]
		r2 := releases[j]
		v1 := r1.Version
		v2 := r2.Version
		if v1 == nil || v2 == nil {
			return r1.Tag <= r2.Tag
		}
		return v1.LessThan(v2)
	})

	for _, release := range releases {
		arr, ok := hrs[release.Tag]
		if !ok {
			continue
		}
		logE.WithField("num_of_assets", len(arr)).Debug("got assets")
		assets := make([]*github.ReleaseAsset, 0, len(arr))
		for _, asset := range arr {
			if excludeAsset(logE, asset, cfg) {
				continue
			}
			assets = append(assets, &github.ReleaseAsset{
				Name: &asset,
			})
		}
		if len(assets) == 0 {
			continue
		}
		release.assets = assets
	}

	return releases
}
