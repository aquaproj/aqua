package versiongetter

import (
	"github.com/antonmedv/expr/vm"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/expr"
)

type Filter struct {
	Prefix     string
	Filter     *vm.Program
	Constraint string
}

func createFilters(pkgInfo *registry.PackageInfo) ([]*Filter, error) {
	filters := make([]*Filter, 0, 1+len(pkgInfo.VersionOverrides))
	topFilter := &Filter{}
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
			Constraint: topFilter.Constraint,
		}
		if vo.VersionFilter != nil {
			f, err := expr.CompileVersionFilter(*vo.VersionFilter)
			if err != nil {
				return nil, err //nolint:wrapcheck
			}
			flt.Filter = f
		}
		flt.Constraint = vo.VersionConstraints
		if vo.VersionPrefix != nil {
			flt.Prefix = *vo.VersionPrefix
		}
		filters = append(filters, flt)
	}
	return filters, nil
}
