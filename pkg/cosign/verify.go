package cosign

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/template"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type VerifierImpl struct {
	executor      Executor
	fs            afero.Fs
	downloader    download.ClientAPI
	cosignExePath string
	disabled      bool
	mutex         *sync.Mutex
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
		mutex:    &sync.Mutex{},
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
		opts = append(opts, "--signature", sigFile.Name())
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

		opts = append(opts, "--key", keyFile.Name())
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
		opts = append(opts, "--certificate", certFile.Name())
	}

	if err := verifier.verify(ctx, logE, &ParamVerify{
		Opts:               opts,
		CosignExperimental: cos.CosignExperimental,
		Target:             verifiedFilePath,
	}); err != nil {
		return fmt.Errorf("verify a signature file with Cosign: %w", logerr.WithFields(err, logrus.Fields{
			"cosign_opts":         strings.Join(opts, ", "),
			"cosign_experimental": cos.CosignExperimental,
			"target":              verifiedFilePath,
		}))
	}
	return nil
}

type Executor interface {
	ExecWithEnvsAndGetCombinedOutput(ctx context.Context, exePath string, args, envs []string) (string, int, error)
}

type ParamVerify struct {
	CosignExperimental bool
	Opts               []string
	Target             string
	CosignExePath      string
}

const tempErrMsg = "resource temporarily unavailable"

var errVerify = errors.New("verify with Cosign")

func (verifier *VerifierImpl) exec(ctx context.Context, args, envs []string) (string, error) {
	// https://github.com/aquaproj/aqua/issues/1555
	verifier.mutex.Lock()
	defer verifier.mutex.Unlock()
	out, _, err := verifier.executor.ExecWithEnvsAndGetCombinedOutput(ctx, verifier.cosignExePath, args, envs)
	return out, err //nolint:wrapcheck
}

func wait(ctx context.Context, logE *logrus.Entry, retryCount int) error {
	randGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))       //nolint:gosec
	waitTime := time.Duration(randGenerator.Intn(1000)) * time.Millisecond //nolint:gomnd
	logE.WithFields(logrus.Fields{
		"retry_count": retryCount,
		"wait_time":   waitTime,
	}).Info("Verification by Cosign failed temporarily, retring")
	if err := util.Wait(ctx, waitTime); err != nil {
		return fmt.Errorf("wait running Cosign: %w", err)
	}
	return nil
}

func (verifier *VerifierImpl) verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify) error {
	envs := []string{}
	args := append([]string{"verify-blob"}, append(param.Opts, param.Target)...)
	for i := 0; i < 5; i++ {
		// https://github.com/aquaproj/aqua/issues/1554
		out, err := verifier.exec(ctx, args, envs)
		if err == nil {
			return nil
		}
		if !strings.Contains(out, tempErrMsg) {
			return fmt.Errorf("verify with cosign: %w", err)
		}
		if i == 4 { //nolint:gomnd
			// skip last wait
			break
		}
		if err := wait(ctx, logE, i+1); err != nil {
			return err
		}
	}
	return errVerify
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
