package installpackage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/minisign"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

type minisignVerifier struct {
	pkg       *config.Package
	installer *DedicatedInstaller
	verifier  MinisignVerifier
	runtime   *runtime.Runtime
	asset     string
	minisign  *registry.Minisign
}

func (s *minisignVerifier) Enabled(logger *slog.Logger) (bool, error) {
	if !s.minisign.GetEnabled() {
		return false, nil
	}

	mPkg := minisign.Package()
	f, err := mPkg.PackageInfo.CheckSupported(s.runtime, s.runtime.Env())
	if err != nil {
		return false, fmt.Errorf("check if minisign supports this environment: %w", err)
	}
	if f {
		return true, nil
	}

	// aqua doesn't manage a minisign binary for this environment, but
	// verification is still possible with a minisign command installed on the system.
	if _, err := minisign.LookSystemExe(); err != nil {
		logger.Warn("minisign doesn't support this environment and minisign isn't found in PATH")
		return false, nil //nolint:nilerr
	}
	logger.Debug("aqua doesn't manage minisign in this environment; using the minisign found in PATH")
	return true, nil
}

func (s *minisignVerifier) Verify(ctx context.Context, logger *slog.Logger, file string) error {
	logger.Info("verify a package with minisign")

	// aqua's managed minisign is only installed for environments it supports.
	// Otherwise verification relies on a minisign command installed on the system.
	supported, err := minisign.Package().PackageInfo.CheckSupported(s.runtime, s.runtime.Env())
	if err != nil {
		return fmt.Errorf("check if minisign supports this environment: %w", err)
	}
	if supported {
		if err := s.installer.install(ctx, logger); err != nil {
			return fmt.Errorf("install minisign: %w", err)
		}
	}

	pkg := s.pkg
	pkgInfo := s.pkg.PackageInfo
	m := s.minisign

	art := pkg.TemplateArtifact(s.runtime, s.asset)

	if err := s.verifier.Verify(ctx, logger, s.runtime, m, art, &download.File{
		RepoOwner: pkgInfo.RepoOwner,
		RepoName:  pkgInfo.RepoName,
		Version:   pkg.Package.Version,
	}, &minisign.ParamVerify{
		ArtifactPath: file,
		PublicKey:    m.PublicKey,
	}); err != nil {
		return fmt.Errorf("verify a package with minisign: %w", err)
	}

	return nil
}
