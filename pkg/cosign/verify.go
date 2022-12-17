package cosign

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/template"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type Verifier struct {
	executor      Executor
	fs            afero.Fs
	downloader    Downloader
	cosignExePath string
}

func NewVerifier(executor Executor, fs afero.Fs, downloader Downloader, param *config.Param, rt *runtime.Runtime) *Verifier {
	return &Verifier{
		executor:   executor,
		fs:         fs,
		downloader: downloader,
		cosignExePath: ExePath(&ParamExePath{
			RootDir: param.RootDir,
			Runtime: rt,
		}),
	}
}

func (verifier *Verifier) SetCosignExePath(p string) {
	verifier.cosignExePath = p
}

type Downloader interface {
	GetReadCloser(ctx context.Context, file *download.File, logE *logrus.Entry) (io.ReadCloser, int64, error)
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

func (verifier *Verifier) verify(ctx context.Context, param *ParamVerify) error {
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

func (verifier *Verifier) Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, file *download.File, cos *registry.Cosign, art *template.Artifact, verifiedFilePath string) error { //nolint:cyclop,funlen
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

		f, err := convertDownloadedFileToFile(cos.Signature, file, rt, art)
		if err != nil {
			return err
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

		f, err := convertDownloadedFileToFile(cos.Key, file, rt, art)
		if err != nil {
			return err
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

		f, err := convertDownloadedFileToFile(cos.Certificate, file, rt, art)
		if err != nil {
			return err
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

func convertDownloadedFileToFile(file *registry.DownloadedFile, art *download.File, rt *runtime.Runtime, tplParam *template.Artifact) (*download.File, error) {
	f := &download.File{
		Type:      file.Type,
		RepoOwner: file.RepoOwner,
		RepoName:  file.RepoName,
		Version:   art.Version,
	}
	switch file.Type {
	case "github_release":
		if f.RepoOwner == "" {
			f.RepoOwner = art.RepoOwner
		}
		if f.RepoName == "" {
			f.RepoName = art.RepoName
		}
		if file.Asset == nil {
			return nil, errors.New("asset is required")
		}
		asset, err := template.Render(*file.Asset, tplParam, rt)
		if err != nil {
			return nil, fmt.Errorf("render an asset template: %w", err)
		}
		f.Asset = asset
		return f, nil
	case "http":
		if file.URL == nil {
			return nil, errors.New("url is required")
		}
		u, err := template.Render(*file.URL, tplParam, rt)
		if err != nil {
			return nil, fmt.Errorf("render a url template: %w", err)
		}
		f.URL = u
		return f, nil
	}
	return nil, logerr.WithFields(errors.New("invalid file type"), logrus.Fields{ //nolint:wrapcheck
		"file_type": file.Type,
	})
}

func (verifier *Verifier) downloadCosignFile(ctx context.Context, logE *logrus.Entry, f *download.File, tf io.Writer) error {
	rc, _, err := verifier.downloader.GetReadCloser(ctx, f, logE)
	if err != nil {
		return fmt.Errorf("get a readcloser: %w", err)
	}
	if _, err := io.Copy(tf, rc); err != nil {
		return fmt.Errorf("download a file: %w", err)
	}
	return nil
}
