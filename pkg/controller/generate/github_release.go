package generate

import (
	"strings"

	"github.com/antonmedv/expr/vm"
	"github.com/aquaproj/aqua/v2/pkg/expr"
)

type Filter struct {
	Prefix     string
	Filter     *vm.Program
	Constraint string
}

func filterTagByFilter(tagName string, filter *Filter) bool {
	sv := tagName
	if filter.Prefix != "" {
		if !strings.HasPrefix(tagName, filter.Prefix) {
			return false
		}
		sv = strings.TrimPrefix(tagName, filter.Prefix)
	}
	if filter.Filter != nil {
		if f, err := expr.EvaluateVersionFilter(filter.Filter, tagName); err != nil || !f {
			return false
		}
	}
	if filter.Constraint == "" {
		return true
	}
	if f, err := expr.EvaluateVersionConstraints(filter.Constraint, tagName, sv); err == nil && f {
		return true
	}
	return false
}
