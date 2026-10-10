package versiongetter

import (
	"log/slog"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/expr-lang/expr/vm"
)

type Filter struct {
	Prefix     string
	Filter     *vm.Program
	Constraint string
	NoAsset    bool
}

// filterTagName says whether a tag is one the package can be installed from.
//
// A prerelease never is: it is published to be tried rather than installed, and a
// version filter that didn't exclude it would have to exclude it again for every
// package. Otherwise the first filter whose constraint the tag meets answers, which is
// what lets a version_override say a range of versions has no asset.
func filterTagName(logger *slog.Logger, tagName string, prerelease bool, filters []*Filter) bool {
	if prerelease {
		return false
	}
	for _, filter := range filters {
		if matchTagByFilter(logger, tagName, filter) {
			return !filter.NoAsset
		}
	}
	return false
}

func createFilters(pkgInfo *registry.PackageInfo) ([]*Filter, error) {
	filters := make([]*Filter, 0, 1+len(pkgInfo.VersionOverrides))
	topFilter := &Filter{
		NoAsset:    pkgInfo.NoAsset || pkgInfo.ErrorMessage != "",
		Prefix:     pkgInfo.VersionPrefix,
		Constraint: pkgInfo.VersionConstraints,
	}
	if pkgInfo.VersionFilter != "" {
		f, err := expr.CompileVersionFilter(pkgInfo.VersionFilter)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
		topFilter.Filter = f
	}
	filters = append(filters, topFilter)

	for _, vo := range pkgInfo.VersionOverrides {
		flt := &Filter{
			Prefix:     topFilter.Prefix,
			Filter:     topFilter.Filter,
			Constraint: vo.VersionConstraints,
			NoAsset:    topFilter.NoAsset,
		}
		if vo.VersionFilter != nil {
			f, err := expr.CompileVersionFilter(*vo.VersionFilter)
			if err != nil {
				return nil, err //nolint:wrapcheck
			}
			flt.Filter = f
		}
		if vo.VersionPrefix != nil {
			flt.Prefix = *vo.VersionPrefix
		}
		if vo.NoAsset != nil {
			flt.NoAsset = *vo.NoAsset
		}
		if vo.ErrorMessage != nil {
			flt.NoAsset = *vo.ErrorMessage != ""
		}
		filters = append(filters, flt)
	}
	return filters, nil
}

func matchTagByFilter(logger *slog.Logger, tagName string, filter *Filter) bool {
	sv := tagName
	if filter.Prefix != "" {
		if !strings.HasPrefix(tagName, filter.Prefix) {
			return false
		}
		sv = strings.TrimPrefix(tagName, filter.Prefix)
	}
	if filter.Filter != nil {
		if f, err := expr.EvaluateVersionFilter(logger, filter.Filter, tagName); err != nil || !f {
			return false
		}
	}
	if filter.Constraint == "" {
		return true
	}
	if f, err := expr.EvaluateVersionConstraints(logger, filter.Constraint, tagName, sv); err == nil && f {
		return true
	}
	return false
}
