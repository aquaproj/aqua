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
	"github.com/aquaproj/aqua/v2/pkg/osexec"
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

// art is used to render the template.
func (v *Verifier) Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, file *download.File, cos *registry.Cosign, art *template.Artifact, verifiedFilePath string) error {
	if v.disabled {
		logE.Debug("verification with cosign is disabled")
		return nil
	}

	opts, err := cos.RenderOpts(rt, art)
	if err != nil {
		return fmt.Errorf("render cosign options: %w", err)
	}

	files := map[string]*registry.DownloadedFile{
		"signature":   cos.Signature,
		"key":         cos.Key,
		"certificate": cos.Certificate,
		"bundle":      cos.Bundle,
	}
	for name, df := range files {
		if df == nil {
			continue
		}
		f, err := v.downloadFile(ctx, logE, rt, file, art, df)
		if f != "" {
			defer v.fs.Remove(f) //nolint:errcheck
		}
		if err != nil {
			return err
		}
		opts = append(opts, "--"+name, f)
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
	ExecStderrAndGetCombinedOutput(cmd *osexec.Cmd) (string, int, error)
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
	cmd := osexec.Command(ctx, v.cosignExePath, args...)
	cmd.Args[0] = "cosign"
	out, _, err := v.executor.ExecStderrAndGetCombinedOutput(cmd)
	return out, err //nolint:wrapcheck
}

func wait(ctx context.Context, logE *logrus.Entry, retryCount int) error {
	randGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))       //nolint:gosec
	waitTime := time.Duration(randGenerator.Intn(1000)) * time.Millisecond //nolint:mnd
	logE.WithFields(logrus.Fields{
		"retry_count": retryCount,
		"wait_time":   waitTime,
	}).Info("Verification by Cosign failed temporarily, retrying")
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
		if i == 4 { //nolint:mnd
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
	defer rc.Close()
	if _, err := io.Copy(tf, rc); err != nil {
		return fmt.Errorf("download a file: %w", err)
	}
	return nil
}

func (v *Verifier) downloadFile(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, file *download.File, art *template.Artifact, downloadedFile *registry.DownloadedFile) (string, error) {
	// --signature cos.Signature - Download a signature file
	sigFile, err := afero.TempFile(v.fs, "", "")
	if err != nil {
		return "", fmt.Errorf("create a temporary file: %w", err)
	}
	fileName := sigFile.Name()

	f, err := download.ConvertDownloadedFileToFile(downloadedFile, file, rt, art)
	if err != nil {
		return fileName, err //nolint:wrapcheck
	}

	if err := v.downloadCosignFile(ctx, logE, f, sigFile); err != nil {
		return fileName, err
	}
	return fileName, nil
}
