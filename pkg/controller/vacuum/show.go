package vacuum

import (
	"context"

	"github.com/sirupsen/logrus"
)

// ListPackages lists the packages based on the provided arguments.
// If the expired flag is set to true, it lists the expired packages.
// Otherwise, it lists all packages.
func (vc *Controller) ListPackages(ctx context.Context, logE *logrus.Entry, expired bool, args ...string) error {
	if expired {
		return vc.handleListExpiredPackages(ctx, logE, args...)
	}
	return vc.handleListPackages(ctx, logE, args...)
}

// handleListPackages retrieves a list of packages and displays them using a fuzzy search.
func (vc *Controller) handleListPackages(ctx context.Context, logE *logrus.Entry, args ...string) error {
	pkgs, err := vc.d.List(ctx, logE)
	if err != nil {
		return err
	}
	return vc.displayPackagesFuzzy(logE, pkgs, args...)
}

// handleListExpiredPackages handles the process of listing expired packages
// and displaying them using a fuzzy search.
func (vc *Controller) handleListExpiredPackages(ctx context.Context, logE *logrus.Entry, args ...string) error {
	expiredPkgs, err := vc.listExpiredPackages(ctx, logE)
	if err != nil {
		return err
	}
	return vc.displayPackagesFuzzy(logE, expiredPkgs, args...)
}
