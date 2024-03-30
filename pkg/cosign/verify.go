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

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
	"github.com/aquaproj/aqua/v2/pkg/timer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

var mutex = &sync.Mutex{} //nolint:gochecknoglobals

func GetMutex() *sync.Mutex {
	return mutex
}

type Verifier struct {
	executor      Executor
	fs            afero.Fs
	downloader    download.ClientAPI
	cosignExePath string
	disabled      bool
}

func NewVerifier(executor Executor, fs afero.Fs, downloader download.ClientAPI, param *config.Param) *Verifier {
	rt := runtime.NewR()
	return &Verifier{
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

func (v *Verifier) Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, file *download.File, cos *registry.Cosign, art *template.Artifact, verifiedFilePath string) error { //nolint:cyclop,funlen
	if v.disabled {
		logE.Debug("verification with cosign is disabled")
		return nil
	}

	opts, err := cos.RenderOpts(rt, art)
	if err != nil {
		return fmt.Errorf("render cosign options: %w", err)
	}

	if cos.Signature != nil {
		sigFile, err := afero.TempFile(v.fs, "", "")
		if err != nil {
			return fmt.Errorf("create a temporal file: %w", err)
		}
		defer v.fs.Remove(sigFile.Name()) //nolint:errcheck

		f, err := download.ConvertDownloadedFileToFile(cos.Signature, file, rt, art)
		if err != nil {
			return err //nolint:wrapcheck
		}

		if err := v.downloadCosignFile(ctx, logE, f, sigFile); err != nil {
			return fmt.Errorf("download a signature: %w", err)
		}
		opts = append(opts, "--signature", sigFile.Name())
	}
	if cos.Key != nil {
		keyFile, err := afero.TempFile(v.fs, "", "")
		if err != nil {
			return fmt.Errorf("create a temporal file: %w", err)
		}
		defer v.fs.Remove(keyFile.Name()) //nolint:errcheck

		f, err := download.ConvertDownloadedFileToFile(cos.Key, file, rt, art)
		if err != nil {
			return err //nolint:wrapcheck
		}

		if err := v.downloadCosignFile(ctx, logE, f, keyFile); err != nil {
			return fmt.Errorf("download a signature: %w", err)
		}

		opts = append(opts, "--key", keyFile.Name())
	}
	if cos.Certificate != nil {
		certFile, err := afero.TempFile(v.fs, "", "")
		if err != nil {
			return fmt.Errorf("create a temporal file: %w", err)
		}
		defer v.fs.Remove(certFile.Name()) //nolint:errcheck

		f, err := download.ConvertDownloadedFileToFile(cos.Certificate, file, rt, art)
		if err != nil {
			return err //nolint:wrapcheck
		}

		if err := v.downloadCosignFile(ctx, logE, f, certFile); err != nil {
			return fmt.Errorf("download a signature: %w", err)
		}

		if err := v.downloadCosignFile(ctx, logE, f, certFile); err != nil {
			return fmt.Errorf("download a certificate: %w", err)
		}
		opts = append(opts, "--certificate", certFile.Name())
	}

	if err := v.verify(ctx, logE, &ParamVerify{
		Opts:   opts,
		Target: verifiedFilePath,
	}); err != nil {
		return fmt.Errorf("verify a signature file with Cosign: %w", logerr.WithFields(err, logrus.Fields{
			"cosign_opts": strings.Join(opts, ", "),
			"target":      verifiedFilePath,
		}))
	}
	return nil
}

type Executor interface {
	ExecWithEnvsAndGetCombinedOutput(ctx context.Context, exePath string, args, envs []string) (string, int, error)
}

type ParamVerify struct {
	Opts          []string
	Target        string
	CosignExePath string
}

var errVerify = errors.New("verify with Cosign")

func (v *Verifier) exec(ctx context.Context, args []string) (string, error) {
	// https://github.com/aquaproj/aqua/issues/1555
	mutex.Lock()
	defer mutex.Unlock()
	out, _, err := v.executor.ExecWithEnvsAndGetCombinedOutput(ctx, v.cosignExePath, args, nil)
	return out, err //nolint:wrapcheck
}

func wait(ctx context.Context, logE *logrus.Entry, retryCount int) error {
	randGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))       //nolint:gosec
	waitTime := time.Duration(randGenerator.Intn(1000)) * time.Millisecond //nolint:gomnd
	logE.WithFields(logrus.Fields{
		"retry_count": retryCount,
		"wait_time":   waitTime,
	}).Info("Verification by Cosign failed temporarily, retring")
	if err := timer.Wait(ctx, waitTime); err != nil {
		return fmt.Errorf("wait running Cosign: %w", err)
	}
	return nil
}

func (v *Verifier) verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify) error {
	args := append([]string{"verify-blob"}, append(param.Opts, param.Target)...)
	for i := range 5 {
		// https://github.com/aquaproj/aqua/issues/1554
		if _, err := v.exec(ctx, args); err == nil {
			return nil
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

func (v *Verifier) downloadCosignFile(ctx context.Context, logE *logrus.Entry, f *download.File, tf io.Writer) error {
	rc, _, err := v.downloader.ReadCloser(ctx, logE, f)
	if err != nil {
		return fmt.Errorf("get a readcloser: %w", err)
	}
	if _, err := io.Copy(tf, rc); err != nil {
		return fmt.Errorf("download a file: %w", err)
	}
	return nil
}
