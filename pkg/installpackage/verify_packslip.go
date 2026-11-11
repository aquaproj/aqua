package installpackage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/packslip"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

type packslipVerifier struct {
	pkg       *config.Package
	installer *DedicatedInstaller
	verifier  PackslipVerifier
	runtime   *runtime.Runtime
	// realRuntime is the host platform. It can be different from runtime,
	// which AQUA_GOOS and AQUA_GOARCH can change.
	realRuntime *runtime.Runtime
	asset       string
	packslip    *registry.Packslip
}

func (*packslipVerifier) Kind() string {
	return "packslip"
}

func (s *packslipVerifier) Enabled(logger *slog.Logger) (bool, error) {
	if !s.packslip.GetEnabled() {
		return false, nil
	}

	// packslip is executed on the host, so its support must be checked with the
	// host platform, not the target platform.
	rt := s.realRuntime
	pPkg := packslip.Package()
	if f, err := pPkg.PackageInfo.CheckSupported(rt, rt.Env()); err != nil {
		return false, fmt.Errorf("check if packslip supports this environment: %w", err)
	} else if !f {
		logger.Warn("packslip doesn't support this environment, so the package isn't verified with packslip",
			"packslip_env", rt.Env())
		return false, nil
	}
	return true, nil
}

func (s *packslipVerifier) Verify(ctx context.Context, logger *slog.Logger, file string) error {
	logger.Info("verify a package with packslip")
	if err := s.installer.install(ctx, logger); err != nil {
		return fmt.Errorf("install packslip: %w", err)
	}

	pkg := s.pkg
	pkgInfo := s.pkg.PackageInfo

	art := pkg.TemplateArtifact(s.runtime, s.asset)

	if err := s.verifier.Verify(ctx, logger, s.runtime, s.packslip, art, &download.File{
		Type:      pkgInfo.Type,
		RepoOwner: pkgInfo.RepoOwner,
		RepoName:  pkgInfo.RepoName,
		Host:      pkgInfo.GetHost(),
		Version:   pkg.Package.Version,
	}, file, s.asset); err != nil {
		return fmt.Errorf("verify a package with packslip: %w", err)
	}

	return nil
}
