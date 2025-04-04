package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/sirupsen/logrus"
)

type slsaVerifier struct {
	disabled  bool
	pkg       *config.Package
	installer *DedicatedInstaller
	verifier  SLSAVerifier
	runtime   *runtime.Runtime
	asset     string
}

func (s *slsaVerifier) Enabled(logE *logrus.Entry) (bool, error) {
	if s.disabled {
		logE.Debug("slsa verification is disabled")
		return false, nil
	}
	return s.pkg.PackageInfo.SLSAProvenance.GetEnabled(), nil
}

func (s *slsaVerifier) Verify(ctx context.Context, logE *logrus.Entry, file string) error {
	logE.Info("verify a package with slsa-verifier")
	if err := s.installer.install(ctx, logE); err != nil {
		return fmt.Errorf("install slsa-verifier: %w", err)
	}

	pkg := s.pkg
	pkgInfo := s.pkg.PackageInfo

	art := pkg.TemplateArtifact(s.runtime, s.asset)

	if err := s.verifier.Verify(ctx, logE, s.runtime, pkgInfo.SLSAProvenance, art, &download.File{
		RepoOwner: pkgInfo.RepoOwner,
		RepoName:  pkgInfo.RepoName,
		Version:   pkg.Package.Version,
	}, &slsa.ParamVerify{
		SourceURI:    pkgInfo.SLSASourceURI(),
		SourceTag:    pkg.Package.Version,
		ArtifactPath: file,
	}); err != nil {
		return fmt.Errorf("verify a package with slsa-verifier: %w", err)
	}
	return nil
}
