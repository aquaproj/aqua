package policy

import (
	"errors"

	"github.com/aquaproj/aqua/pkg/runtime"
)

var (
	errUnAllowedPackage  = errors.New("this package isn't allowed")
	errUnAllowedRegistry = errors.New("this registry isn't allowed")
)

type Checker struct {
	rt *runtime.Runtime
}

func NewChecker(rt *runtime.Runtime) *Checker {
	return &Checker{
		rt: rt,
	}
}
