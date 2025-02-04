package versiongetter

import (
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

func createFilters(pkgInfo *registry.PackageInfo) ([]*Filter, error) {
	filters := make([]*Filter, 0, 1+len(pkgInfo.VersionOverrides))
	topFilter := &Filter{
		NoAsset: pkgInfo.NoAsset,
	}
	if pkgInfo.ErrorMessage != "" {
		pkgInfo.NoAsset = true
	}
	if pkgInfo.VersionFilter != "" {
		f, err := expr.CompileVersionFilter(pkgInfo.VersionFilter)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
		topFilter.Filter = f
	}
	topFilter.Constraint = pkgInfo.VersionConstraints
	if pkgInfo.VersionPrefix != "" {
		topFilter.Prefix = pkgInfo.VersionPrefix
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
		if vo.NoAsset != nil && *vo.NoAsset {
			flt.NoAsset = true
		}
		if vo.ErrorMessage != nil && *vo.ErrorMessage != "" {
			flt.NoAsset = true
		}
		filters = append(filters, flt)
	}
	return filters, nil
}
