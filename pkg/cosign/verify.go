package cosign

import (
	"context"
	"fmt"
	"io"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/template"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type VerifierImpl struct {
	executor      Executor
	fs            afero.Fs
	downloader    download.ClientAPI
	cosignExePath string
	disabled      bool
}

func NewVerifier(executor Executor, fs afero.Fs, downloader download.ClientAPI, param *config.Param) *VerifierImpl {
	rt := runtime.NewR()
	return &VerifierImpl{
		executor:   executor,
		fs:         fs,
		downloader: downloader,
		cosignExePath: ExePath(&ParamExePath{
			RootDir: param.RootDir,
			Runtime: rt,
		}),
		// assets for windows/arm64 aren't released.
		disabled: rt.GOOS == "windows" && rt.GOARCH == "arm64",
	}
}

type Verifier interface {
	Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, file *download.File, cos *registry.Cosign, art *template.Artifact, verifiedFilePath string) error
}

type MockVerifier struct {
	err error
}

func (mock *MockVerifier) Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, file *download.File, cos *registry.Cosign, art *template.Artifact, verifiedFilePath string) error {
	return mock.err
}

func (verifier *VerifierImpl) Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, file *download.File, cos *registry.Cosign, art *template.Artifact, verifiedFilePath string) error { //nolint:cyclop,funlen
	if verifier.disabled {
		logE.Debug("verification with cosign is disabled")
		return nil
	}
	opts, err := cos.RenderOpts(rt, art)
	if err != nil {
		return fmt.Errorf("render cosign options: %w", err)
	}
	cos.Opts = opts

	if cos.Signature != nil {
		sigFile, err := afero.TempFile(verifier.fs, "", "")
		if err != nil {
			return fmt.Errorf("create a temporal file: %w", err)
		}
		defer verifier.fs.Remove(sigFile.Name()) //nolint:errcheck

		f, err := download.ConvertDownloadedFileToFile(cos.Signature, file, rt, art)
		if err != nil {
			return err //nolint:wrapcheck
		}

		if err := verifier.downloadCosignFile(ctx, logE, f, sigFile); err != nil {
			return fmt.Errorf("download a signature: %w", err)
		}
		cos.Opts = append(cos.Opts, "--signature", sigFile.Name())
	}
	if cos.Key != nil {
		keyFile, err := afero.TempFile(verifier.fs, "", "")
		if err != nil {
			return fmt.Errorf("create a temporal file: %w", err)
		}
		defer verifier.fs.Remove(keyFile.Name()) //nolint:errcheck

		f, err := download.ConvertDownloadedFileToFile(cos.Key, file, rt, art)
		if err != nil {
			return err //nolint:wrapcheck
		}

		if err := verifier.downloadCosignFile(ctx, logE, f, keyFile); err != nil {
			return fmt.Errorf("download a signature: %w", err)
		}

		cos.Opts = append(cos.Opts, "--key", keyFile.Name())
	}
	if cos.Certificate != nil {
		certFile, err := afero.TempFile(verifier.fs, "", "")
		if err != nil {
			return fmt.Errorf("create a temporal file: %w", err)
		}
		defer verifier.fs.Remove(certFile.Name()) //nolint:errcheck

		f, err := download.ConvertDownloadedFileToFile(cos.Certificate, file, rt, art)
		if err != nil {
			return err //nolint:wrapcheck
		}

		if err := verifier.downloadCosignFile(ctx, logE, f, certFile); err != nil {
			return fmt.Errorf("download a signature: %w", err)
		}

		if err := verifier.downloadCosignFile(ctx, logE, f, certFile); err != nil {
			return fmt.Errorf("download a certificate: %w", err)
		}
		cos.Opts = append(cos.Opts, "--certificate", certFile.Name())
	}

	if err := verifier.verify(ctx, &ParamVerify{
		Opts:               cos.Opts,
		CosignExperimental: cos.CosignExperimental,
		Target:             verifiedFilePath,
	}); err != nil {
		return fmt.Errorf("verify a signature file with Cosign: %w", err)
	}
	return nil
}

type Executor interface {
	ExecWithEnvs(ctx context.Context, exePath string, args, envs []string) (int, error)
}

type ParamVerify struct {
	CosignExperimental bool
	Opts               []string
	Target             string
	CosignExePath      string
}

func (verifier *VerifierImpl) verify(ctx context.Context, param *ParamVerify) error {
	envs := []string{}
	if param.CosignExperimental {
		envs = []string{"COSIGN_EXPERIMENTAL=1"}
	}
	_, err := verifier.executor.ExecWithEnvs(ctx, verifier.cosignExePath, append([]string{"verify-blob"}, append(param.Opts, param.Target)...), envs)
	if err != nil {
		return fmt.Errorf("verify with cosign: %w", err)
	}
	return nil
}

func (verifier *VerifierImpl) downloadCosignFile(ctx context.Context, logE *logrus.Entry, f *download.File, tf io.Writer) error {
	rc, _, err := verifier.downloader.GetReadCloser(ctx, logE, f)
	if err != nil {
		return fmt.Errorf("get a readcloser: %w", err)
	}
	if _, err := io.Copy(tf, rc); err != nil {
		return fmt.Errorf("download a file: %w", err)
	}
	return nil
}
