package domain

import (
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/security"
)

type SecurityChecker interface {
	Validate(pkg *config.Package, secConfig *security.Config) error
}
