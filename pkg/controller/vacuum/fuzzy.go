package vacuum

import (
	"errors"
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/sirupsen/logrus"
)

func (vc *Controller) displayPackagesFuzzyTest(logE *logrus.Entry, pkgs []*PackageVacuumEntry) error {
	var pkgInformations struct {
		TotalPackages int
		TotalExpired  int
	}
	for _, pkg := range pkgs {
		if vc.isPackageExpired(pkg) {
			pkgInformations.TotalExpired++
		}
		pkgInformations.TotalPackages++
	}
	// Display log entry with informations for testing purposes
	logE.WithFields(logrus.Fields{
		"TotalPackages": pkgInformations.TotalPackages,
		"TotalExpired":  pkgInformations.TotalExpired,
	}).Info("Test mode: Displaying packages")
	return nil
}

func (vc *Controller) displayPackagesFuzzy(logE *logrus.Entry, pkgs []*PackageVacuumEntry, args ...string) error {
	if len(pkgs) == 0 {
		logE.Info("no packages to display")
		return nil
	}
	if len(args) > 0 && args[0] == "test" {
		return vc.displayPackagesFuzzyTest(logE, pkgs)
	}
	return vc.displayPackagesFuzzyInteractive(pkgs)
}

func (vc *Controller) displayPackagesFuzzyInteractive(pkgs []*PackageVacuumEntry) error {
	_, err := fuzzyfinder.Find(pkgs, func(i int) string {
		var expiredString string
		if vc.isPackageExpired(pkgs[i]) {
			expiredString = "⌛ "
		}

		return fmt.Sprintf("%s%s [%s]",
			expiredString,
			pkgs[i].PkgPath,
			humanize.Time(pkgs[i].PackageEntry.LastUsageTime),
		)
	},
		fuzzyfinder.WithPreviewWindow(func(i, _, _ int) string {
			if i == -1 {
				return "No package selected"
			}
			pkg := pkgs[i]
			var expiredString string
			if vc.isPackageExpired(pkg) {
				expiredString = "Expired ⌛"
			}
			return fmt.Sprintf(
				"Package Details:\n\n"+
					"%s \n"+
					"Type: %s\n"+
					"Package: %s\n"+
					"Version: %s\n\n"+
					"Last Used: %s\n"+
					"Last Used (exact): %s\n\n",
				expiredString,
				pkg.PackageEntry.Package.Type,
				pkg.PackageEntry.Package.Name,
				pkg.PackageEntry.Package.Version,
				humanize.Time(pkg.PackageEntry.LastUsageTime),
				pkg.PackageEntry.LastUsageTime.Format("2006-01-02 15:04:05"),
			)
		}),
		fuzzyfinder.WithHeader("Navigate through packages to display details"),
	)
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil
		}
		return fmt.Errorf("display packages: %w", err)
	}

	return nil
}
