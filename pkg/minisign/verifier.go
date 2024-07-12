package minisign

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

type Verifier struct {
	downloader download.ClientAPI
	fs         afero.Fs
	exe        Executor
}

func New(downloader download.ClientAPI, fs afero.Fs, exe Executor) *Verifier {
	return &Verifier{
		downloader: downloader,
		fs:         fs,
		exe:        exe,
	}
}

type ParamVerify struct {
	ArtifactPath string
	PublicKey    string
}

func (v *Verifier) Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, m *registry.Minisign, art *template.Artifact, file *download.File, param *ParamVerify) error {
	f, err := download.ConvertDownloadedFileToFile(m.ToDownloadedFile(), file, rt, art)
	if err != nil {
		return err //nolint:wrapcheck
	}
	rc, _, err := v.downloader.ReadCloser(ctx, logE, f)
	if err != nil {
		return fmt.Errorf("download a Minisign signature: %w", err)
	}
	defer rc.Close()

	signatureFile, err := afero.TempFile(v.fs, "", "")
	if err != nil {
		return fmt.Errorf("create a temporal file: %w", err)
	}
	defer signatureFile.Close()
	defer v.fs.Remove(signatureFile.Name()) //nolint:errcheck
	if _, err := io.Copy(signatureFile, rc); err != nil {
		return fmt.Errorf("copy a signature to a temporary file: %w", err)
	}

	return v.exe.Verify(ctx, logE, param, signatureFile.Name()) //nolint:wrapcheck
}
