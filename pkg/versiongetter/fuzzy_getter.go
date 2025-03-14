package versiongetter

import (
	"context"
	"strings"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type FuzzyGetter struct {
	fuzzyFinder FuzzyFinder
	getter      VersionGetter
}

func NewFuzzy(finder FuzzyFinder, getter VersionGetter) *FuzzyGetter {
	return &FuzzyGetter{
		fuzzyFinder: finder,
		getter:      getter,
	}
}

type FuzzyFinder interface {
	Find(items []*fuzzyfinder.Item, hasPreview bool) (int, error)
	FindMulti(items []*fuzzyfinder.Item, hasPreview bool) ([]int, error)
}

func (g *FuzzyGetter) Get(ctx context.Context, logE *logrus.Entry, pkg *registry.PackageInfo, currentVersion string, useFinder bool, limit int) string { //nolint:cyclop
	filters, err := createFilters(pkg)
	if err != nil {
		logerr.WithError(logE, err).Warn("create filters")
		return ""
	}

	repoName := pkg.RepoOwner + "/" + pkg.RepoName
	logE = logE.WithField("repository", repoName)
	if useFinder { //nolint:nestif
		logE := logE.WithFields(nil) // Copy logE because g.getter.List has a side effect to change logE
		start := time.Now()
		versions, err := g.getter.List(ctx, logE, pkg, filters, limit)
		elapsed := time.Since(start)
		if err != nil {
			logE.WithError(err).Warn("retrieve package versions")
			return ""
		}
		if versions == nil {
			return ""
		}
		currentVersionIndex := 0
		if currentVersion != "" {
			for i, version := range versions {
				if version.Item == currentVersion {
					version.Item += " (*)"
					currentVersionIndex = i
					break
				}
			}
		}
		idx, err := g.fuzzyFinder.Find(versions, true)
		logE.Debug("retrieve package versions in ", elapsed) // finder's output will overwrite log, so log after it
		if err != nil {
			return ""
		}
		if idx == currentVersionIndex {
			return strings.TrimSuffix(versions[idx].Item, " (*)")
		}
		return versions[idx].Item
	}

	start := time.Now()
	version, err := g.getter.Get(ctx, logE, pkg, filters)
	logE.Debug("retrieve package versions in ", time.Since(start))
	if err != nil {
		logerr.WithError(logE, err).Warn("retrieve package versions")
		return ""
	}
	return version
}
