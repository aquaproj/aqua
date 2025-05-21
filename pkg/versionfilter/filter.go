package versionfilter

import (
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/expr-lang/expr/vm"
)

type RawVersion interface {
	// GitHub Release
	// GitHub Tag
	// Cargo
	Version() string
}

type Version struct {
	Semver  string
	Version string
	Prefix  string
}

type Filter interface {
	Filter(v string) (bool, error)
}

type ConbinedFilter struct {
	filters []Filter
}

func Combine(filters ...Filter) *ConbinedFilter {
	return &ConbinedFilter{filters: filters}
}

type InverseFilter struct {
	filter Filter
}

func Inverse(f Filter) *InverseFilter {
	return &InverseFilter{filter: f}
}

func (f *InverseFilter) Filter(v string) (bool, error) {
	a, err := f.filter.Filter(v)
	if err != nil {
		return false, err
	}
	return !a, nil
}

func (cf *ConbinedFilter) Filter(v string) (bool, error) {
	for _, f := range cf.filters {
		ok, err := f.Filter(v)
		if err != nil {
			continue
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

type PrefixFilter struct {
	prefix string
}

func NewPrefixFilter(prefix string) *PrefixFilter {
	return &PrefixFilter{prefix: prefix}
}

func (f *PrefixFilter) Filter(v string) (bool, error) {
	return strings.HasPrefix(v, f.prefix), nil
}

type ExprFilter struct {
	expr    *vm.Program
	inverse bool
}

func NewExprFilter(expr *vm.Program) *ExprFilter {
	return &ExprFilter{
		expr: expr,
	}
}

func (f *ExprFilter) Filter(v string) (bool, error) {
	return expr.EvaluateVersionFilter(f.expr, v)
}

/*
version_prefix: cli-
version_filter: semver(">=1.0.0")
error_message: "this package is not supported"
no_asset: true
version_constraint: semver(">=1.0.0")
*/

func createFilterFromPkgInfo(pkgInfo *registry.PackageInfo) (Filter, error) {
	var arr []Filter
	if pkgInfo.VersionFilter != "" {
		f, err := expr.CompileVersionFilter(pkgInfo.VersionFilter)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
		arr = append(arr, NewExprFilter(f))
	}
	if pkgInfo.VersionPrefix != "" {
		arr = append(arr, NewPrefixFilter(pkgInfo.VersionPrefix))
	}
	for _, vo := range pkgInfo.VersionOverrides {
		var filter Filter
		if vo.VersionFilter != nil {
			f, err := expr.CompileVersionFilter(*vo.VersionFilter)
			if err != nil {
				return nil, err //nolint:wrapcheck
			}
			filter = NewExprFilter(f)
		}
		if vo.VersionPrefix != nil {
			filter = NewPrefixFilter(*vo.VersionPrefix)
		}
		if vo.Excluded() {
			arr = append(arr, Inverse(filter))
		} else {
			arr = append(arr, filter)
		}
		flt := &Filter{
			Prefix:     topFilter.Prefix,
			Filter:     topFilter.Filter,
			Constraint: vo.VersionConstraints,
			NoAsset:    topFilter.NoAsset,
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
	return Combine(arr...), nil
}
