package config

import (
	"errors"

	"github.com/aquaproj/aqua/pkg/runtime"
)

type PolicyChecker struct {
	rt *runtime.Runtime
}

func NewPolicyChecker(rt *runtime.Runtime) *PolicyChecker {
	return &PolicyChecker{
		rt: rt,
	}
}

var (
	errUnAllowedPackage  = errors.New("this package isn't allowed")
	errUnAllowedRegistry = errors.New("this registry isn't allowed")
)
