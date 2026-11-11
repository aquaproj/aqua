package cosign

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/exec"
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
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

var mutex = &sync.Mutex{} //nolint:gochecknoglobals

func GetMutex() *sync.Mutex {
	return mutex
}

type Verifier struct {
	executor      Executor
	downloader    download.ClientAPI
	cosignExePath string
	disabled      bool
}

func NewVerifier(executor Executor, downloader download.ClientAPI, param *config.Param) *Verifier {
	rt := runtime.NewR(context.Background())
	return &Verifier{
		executor:   executor,
		downloader: downloader,
		cosignExePath: ExePath(&ParamExePath{
			RootDir: param.RootDir,
			Runtime: rt,
		}),
		// assets for windows/arm64 aren't released.
		disabled: rt.GOOS == "windows" && rt.GOARCH == "arm64",
	}
}

func (v *Verifier) Verify(ctx context.Context, logger *slog.Logger, rt *runtime.Runtime, file *download.File, cos *registry.Cosign, art *template.Artifact, verifiedFilePath string) error {
	// art is used to render the template.
	if v.disabled {
		logger.Debug("verification with cosign is disabled")
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
		f, err := v.downloadFile(ctx, logger, rt, file, art, df)
		if f != "" {
			defer os.Remove(f)
		}
		if err != nil {
			return err
		}
		opts = append(opts, "--"+name, f)
	}

	if err := v.verify(ctx, logger, &ParamVerify{
		Opts:   opts,
		Target: verifiedFilePath,
	}); err != nil {
		return fmt.Errorf("verify a signature file with Cosign: %w", slogerr.With(err,
			"cosign_opts", strings.Join(opts, ", "),
			"target", verifiedFilePath,
		))
	}
	return nil
}

// Supported reports whether cosign runs on this host. Where it doesn't, Verify
// accepts everything without looking, so a caller recording what was verified has
// to ask first.
func (v *Verifier) Supported() bool {
	return !v.disabled
}

// VerifyDigest verifies the signature against the asset's digest instead of its
// bytes, so that the asset needn't be downloaded. digest is "<algorithm>:<hex>".
//
// cosign verify-blob takes a digest in place of the blob only for a bundle in the
// new format (https://github.com/aquaproj/aqua/issues/5280): given an old bundle, or
// --signature and --certificate, it tries to open the digest as a file. So the
// bundle is downloaded and looked at first, and false is returned without running
// cosign when the digest can't be used; the caller then verifies the asset itself.
func (v *Verifier) VerifyDigest(ctx context.Context, logger *slog.Logger, rt *runtime.Runtime, file *download.File, cos *registry.Cosign, art *template.Artifact, digest string) (bool, error) {
	if v.disabled {
		return false, nil
	}
	opts, err := cos.RenderOpts(rt, art)
	if err != nil {
		return false, fmt.Errorf("render cosign options: %w", err)
	}
	if !digestUsable(cos, opts) {
		return false, nil
	}
	bundle, err := v.downloadFile(ctx, logger, rt, file, art, cos.Bundle)
	if bundle != "" {
		defer os.Remove(bundle)
	}
	if err != nil {
		return false, err
	}
	if f, err := isNewBundle(bundle); err != nil || !f {
		if err == nil {
			logger.Debug("the cosign bundle is in the old format, so the signature can't be verified with the digest")
		}
		return false, err
	}
	opts = append(opts, "--bundle", bundle)
	if err := v.verify(ctx, logger, &ParamVerify{
		Opts:   opts,
		Target: digest,
	}); err != nil {
		return false, fmt.Errorf("verify a digest with Cosign: %w", slogerr.With(err,
			"cosign_opts", strings.Join(opts, ", "),
			"digest", digest,
		))
	}
	return true, nil
}

// digestUsable reports whether the configuration verifies with a bundle alone, which
// is the only way cosign may verify a digest.
func digestUsable(cos *registry.Cosign, opts []string) bool {
	return cos.Bundle != nil && cos.Signature == nil && cos.Certificate == nil && cos.Key == nil && !hasFileOpt(opts)
}

// hasFileOpt reports whether opts give cosign a signature, certificate, key or bundle
// of their own, which the digest path can't be sure of.
func hasFileOpt(opts []string) bool {
	for _, opt := range opts {
		name, _, _ := strings.Cut(opt, "=")
		switch name {
		case "--signature", "--certificate", "--cert", "--key", "--bundle":
			return true
		}
	}
	return false
}

// isNewBundle reports whether the file is a Sigstore bundle in the format that
// carries a mediaType, the only one cosign verifies a digest with.
func isNewBundle(p string) (bool, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return false, fmt.Errorf("read a cosign bundle: %w", err)
	}
	bundle := &struct {
		MediaType string `json:"mediaType"`
	}{}
	if err := json.Unmarshal(b, bundle); err != nil {
		// Not JSON at all is not a bundle cosign takes a digest with either.
		return false, nil //nolint:nilerr
	}
	return strings.HasPrefix(bundle.MediaType, "application/vnd.dev.sigstore.bundle"), nil
}

type Executor interface {
	ExecStderrAndGetCombinedOutput(cmd *osexec.Cmd) (string, int, error)
}

type ParamVerify struct {
	Opts          []string
	Target        string
	CosignExePath string
}

// ErrVerify says the tool ran and didn't accept what it was given.
//
// It is what tells a caller apart from the failures that happen before anything is
// verified -- a signature that isn't published, a file that can't be downloaded --
// which say the package isn't signed the way it was thought to be rather than that
// the signature doesn't hold. What the tool printed is wrapped with it.
var ErrVerify = errors.New("verify with Cosign")

func (v *Verifier) exec(ctx context.Context, args []string) (string, error) {
	// https://github.com/aquaproj/aqua/issues/1555
	mutex.Lock()
	defer mutex.Unlock()
	cmd := osexec.Command(ctx, v.cosignExePath, args...)
	cmd.Args[0] = "cosign"
	out, _, err := v.executor.ExecStderrAndGetCombinedOutput(cmd)
	return out, err //nolint:wrapcheck
}

// ran reports whether the verifier ran and exited, as opposed to never starting.
//
// An exit status is a verdict on what it was given. Anything else -- the executable
// isn't there, the context ended -- happened before the verifier looked at anything,
// and saying the signature didn't hold would be answering a question nobody got to
// ask.
func ran(err error) bool {
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr)
}

// maxWait caps the backoff. The doubling reaches 16 seconds in the five attempts
// cosign is given, so nothing here meets the cap; it is what keeps raising that
// number from turning a retry into an outage.
const maxWait = 30 * time.Second

// waitTime is how long to wait before the next attempt: 2, 4, 8 and 16 seconds, each
// with up to a second of jitter on top.
//
// The wait doubles because what it is waiting out is usually the transparency log
// asking to be left alone for a moment, and five tries a few hundred milliseconds
// apart is over before that has passed. The jitter keeps a machine verifying many
// assets from retrying all of them on the same beat.
func waitTime(retryCount int) time.Duration {
	// 1<<retryCount is two to the power of retryCount, and retryCount runs from 1
	// to 4.
	return min(time.Duration(1<<retryCount)*time.Second, maxWait) + time.Duration(rand.IntN(1000))*time.Millisecond //nolint:gosec,mnd
}

// wait says what it is doing and then does it.
func wait(ctx context.Context, logger *slog.Logger, retryCount int) error {
	wait := waitTime(retryCount)
	logger.Info("Verification by Cosign failed temporarily, retrying",
		"retry_count", retryCount,
		"wait_time", wait)
	if err := timer.Wait(ctx, wait); err != nil {
		return fmt.Errorf("wait running Cosign: %w", err)
	}
	return nil
}

func (v *Verifier) verify(ctx context.Context, logger *slog.Logger, param *ParamVerify) error {
	args := append([]string{"verify-blob"}, append(param.Opts, param.Target)...)
	out := ""
	for i := range 5 {
		// https://github.com/aquaproj/aqua/issues/1554
		o, err := v.exec(ctx, args)
		if err == nil {
			return nil
		}
		if !ran(err) {
			return fmt.Errorf("run cosign: %w", err)
		}
		out = o
		if i == 4 { //nolint:mnd
			// skip last wait
			break
		}
		if err := wait(ctx, logger, i+1); err != nil {
			return err
		}
	}
	// What cosign said, because the reasons differ in what to do about them: a
	// signature that doesn't match is a package to stop installing, and a transient
	// failure reaching the transparency log is a command to run again. Without it
	// both arrive as the same sentence.
	if out = strings.TrimSpace(out); out != "" {
		return fmt.Errorf("%w: %s", ErrVerify, out)
	}
	return ErrVerify
}

func (v *Verifier) downloadCosignFile(ctx context.Context, logger *slog.Logger, f *download.File, tf io.Writer) error {
	rc, _, err := v.downloader.ReadCloser(ctx, logger, f)
	if err != nil {
		return fmt.Errorf("get a readcloser: %w", err)
	}
	defer rc.Close()
	if _, err := io.Copy(tf, rc); err != nil {
		return fmt.Errorf("download a file: %w", err)
	}
	return nil
}

func (v *Verifier) downloadFile(ctx context.Context, logger *slog.Logger, rt *runtime.Runtime, file *download.File, art *template.Artifact, downloadedFile *registry.DownloadedFile) (string, error) {
	// --signature cos.Signature - Download a signature file
	sigFile, err := os.CreateTemp("", "")
	if err != nil {
		return "", fmt.Errorf("create a temporary file: %w", err)
	}
	fileName := sigFile.Name()

	f, err := download.ConvertDownloadedFileToFile(downloadedFile, file, rt, art)
	if err != nil {
		return fileName, err //nolint:wrapcheck
	}

	if err := v.downloadCosignFile(ctx, logger, f, sigFile); err != nil {
		return fileName, err
	}
	return fileName, nil
}
