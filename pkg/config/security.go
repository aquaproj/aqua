package config

import (
	"errors"

	"github.com/aquaproj/aqua/pkg/config/security"
	"github.com/aquaproj/aqua/pkg/runtime"
)

type SecurityChecker struct {
	rt *runtime.Runtime
}

func NewSecurityChecker(rt *runtime.Runtime) *SecurityChecker {
	return &SecurityChecker{
		rt: rt,
	}
}

var (
	errUnAllowedPackage  = errors.New("this package isn't allowed")
	errUnAllowedRegistry = errors.New("this registry isn't allowed")
)

func (sec *SecurityChecker) Validate(pkg *Package, secConfig *security.Config) error {
	if secConfig == nil {
		return errUnAllowedPackage
	}
	if err := sec.validateRegistries(pkg, secConfig.Registries); err != nil {
		return err
	}
	if err := sec.validatePkgs(pkg, secConfig.Packages); err != nil {
		return err
	}
	return nil
}
