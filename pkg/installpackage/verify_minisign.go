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
	// realRuntime is the host platform. It can be different from runtime,
	// which AQUA_GOOS and AQUA_GOARCH can change.
	realRuntime *runtime.Runtime
	asset       string
	minisign    *registry.Minisign
}

func (s *minisignVerifier) Enabled(logger *slog.Logger) (bool, error) {
	if !s.minisign.GetEnabled() {
		return false, nil
	}

	// minisign is executed on the host, so its support must be checked with the
	// host platform, not the target platform.
	rt := s.realRuntime
	mPkg := minisign.Package()
	if f, err := mPkg.PackageInfo.CheckSupported(rt, rt.Env()); err != nil {
		return false, fmt.Errorf("check if minisign supports this environment: %w", err)
	} else if !f {
		logger.Warn("minisign doesn't support this environment, so the package isn't verified with minisign",
			"minisign_env", rt.Env())
		return false, nil
	}
	return true, nil
}

func (s *minisignVerifier) Verify(ctx context.Context, logger *slog.Logger, file string) error {
	logger.Info("verify a package with minisign")
	if err := s.installer.install(ctx, logger); err != nil {
		return fmt.Errorf("install minisign: %w", err)
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
