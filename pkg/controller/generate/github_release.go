package generate

import (
	"github.com/antonmedv/expr/vm"
)

type Filter struct {
	Prefix     string
	Filter     *vm.Program
	Constraint string
}
