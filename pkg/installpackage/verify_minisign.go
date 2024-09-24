package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/minisign"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
)

type minisignVerifier struct {
	pkg       *config.Package
	installer *DedicatedInstaller
	verifier  MinisignVerifier
	runtime   *runtime.Runtime
	asset     string
	minisign  *registry.Minisign
}

func (s *minisignVerifier) Enabled(logE *logrus.Entry) (bool, error) {
	if !s.minisign.GetEnabled() {
		return false, nil
	}

	mPkg := minisign.Package()
	if f, err := mPkg.PackageInfo.CheckSupported(s.runtime, s.runtime.Env()); err != nil {
		return false, fmt.Errorf("check if minisign supports this environment: %w", err)
	} else if !f {
		logE.Warn("minisign doesn't support this environment")
		return false, nil
	}
	return true, nil
}

func (s *minisignVerifier) Verify(ctx context.Context, logE *logrus.Entry, file string) error {
	logE.Info("verify a package with minisign")
	if err := s.installer.install(ctx, logE); err != nil {
		return fmt.Errorf("install minisign: %w", err)
	}

	pkg := s.pkg
	pkgInfo := s.pkg.PackageInfo
	m := s.minisign

	art := pkg.TemplateArtifact(s.runtime, s.asset)

	if err := s.verifier.Verify(ctx, logE, s.runtime, m, art, &download.File{
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
