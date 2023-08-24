package slsa

import (
	"context"
	"fmt"
	"io"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type VerifierImpl struct {
	downloader download.ClientAPI
	fs         afero.Fs
	exe        Executor
}

func New(downloader download.ClientAPI, fs afero.Fs, exe Executor) *VerifierImpl {
	return &VerifierImpl{
		downloader: downloader,
		fs:         fs,
		exe:        exe,
	}
}

type Verifier interface {
	Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, sp *registry.SLSAProvenance, art *template.Artifact, file *download.File, param *ParamVerify) error
}

type MockVerifier struct {
	err error
}

func (m *MockVerifier) Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, sp *registry.SLSAProvenance, art *template.Artifact, file *download.File, param *ParamVerify) error {
	return m.err
}

type ParamVerify struct {
	// e.g. github.com/suzuki-shunsuke/test-cosign-keyless-aqua
	SourceURI string
	// e.g. v0.1.0-7
	SourceTag    string
	ArtifactPath string
}

func (v *VerifierImpl) Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, sp *registry.SLSAProvenance, art *template.Artifact, file *download.File, param *ParamVerify) error {
	f, err := download.ConvertDownloadedFileToFile(sp.ToDownloadedFile(), file, rt, art)
	if err != nil {
		return err //nolint:wrapcheck
	}
	rc, _, err := v.downloader.GetReadCloser(ctx, logE, f)
	if err != nil {
		return fmt.Errorf("download a SLSA Provenance: %w", err)
	}
	defer rc.Close()

	provenanceFile, err := afero.TempFile(v.fs, "", "")
	if err != nil {
		return fmt.Errorf("create a temporal file: %w", err)
	}
	defer provenanceFile.Close()
	defer v.fs.Remove(provenanceFile.Name()) //nolint:errcheck
	if _, err := io.Copy(provenanceFile, rc); err != nil {
		return fmt.Errorf("copy a provenance to a temporal file: %w", err)
	}

	return v.exe.Verify(ctx, logE, param, provenanceFile.Name()) //nolint:wrapcheck
}
