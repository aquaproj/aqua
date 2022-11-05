package domain

import (
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/policy"
)

type PolicyChecker interface {
	Validate(pkg *config.Package, policyConfig *policy.Config) error
}
