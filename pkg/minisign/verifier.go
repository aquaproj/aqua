package minisign

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
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

func (v *Verifier) Verify(ctx context.Context, logger *slog.Logger, rt *runtime.Runtime, m *registry.Minisign, art *template.Artifact, file *download.File, param *ParamVerify) error {
	sigFile, err := v.downloadSignature(ctx, logger, rt, m, art, file)
	if err != nil {
		return err
	}
	defer v.fs.Remove(sigFile)                       //nolint:errcheck
	return v.exe.Verify(ctx, logger, param, sigFile) //nolint:wrapcheck
}

func (v *Verifier) downloadSignature(ctx context.Context, logger *slog.Logger, rt *runtime.Runtime, m *registry.Minisign, art *template.Artifact, file *download.File) (string, error) {
	f, err := download.ConvertDownloadedFileToFile(m.ToDownloadedFile(), file, rt, art)
	if err != nil {
		return "", err //nolint:wrapcheck
	}

	rc, _, err := v.downloader.ReadCloser(ctx, logger, f)
	if err != nil {
		return "", fmt.Errorf("download a Minisign signature: %w", err)
	}
	defer rc.Close()

	signatureFile, err := afero.TempFile(v.fs, "", "")
	if err != nil {
		return "", fmt.Errorf("create a temporary file: %w", err)
	}
	defer signatureFile.Close()
	if _, err := io.Copy(signatureFile, rc); err != nil {
		return signatureFile.Name(), fmt.Errorf("copy a signature to a temporary file: %w", err)
	}
	return signatureFile.Name(), nil
}
