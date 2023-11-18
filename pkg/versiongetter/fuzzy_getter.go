package versiongetter

import (
	"context"
	"strings"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/sirupsen/logrus"
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
		return ""
	}

	const ( // log message
		getVerTimeInfo = "Retrieve pkg version(s) in "
		getVerErrWarn  = "Version retrieving error"
	)

	repoName := pkg.RepoOwner + "/" + pkg.RepoName
	logE = logE.WithField("repo", repoName)
	start := time.Now()
	if useFinder { //nolint:nestif
		logE := logE.WithFields(nil) // Copy logE becuse g.getter.List has a side effect to change logE
		versions, err := g.getter.List(ctx, logE, pkg, filters, limit)
		elapsed := time.Since(start)
		if err != nil {
			logE.WithError(err).Warn(getVerErrWarn)
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
		logE.Debug(getVerTimeInfo, elapsed) // finder's output will overwrite log, so log after it
		if err != nil {
			return ""
		}
		if idx == currentVersionIndex {
			return strings.TrimSuffix(versions[idx].Item, " (*)")
		}
		return versions[idx].Item
	}

	version, err := g.getter.Get(ctx, logE, pkg, filters)
	logE.Debug(getVerTimeInfo, time.Since(start))
	if err != nil {
		logE.WithError(err).Warn(getVerErrWarn)
		return ""
	}
	return version
}
