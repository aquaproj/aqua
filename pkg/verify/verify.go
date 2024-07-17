package verify

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/timer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type Tool interface {
	Package() *config.Package
	Checksums() map[string]string
	Enabled(pkg *registry.PackageInfo) bool
	SupportedConfig() bool
	Signature(ctx context.Context, logE *logrus.Entry) (*registry.DownloadedFile, string, error)
	Command(verifiedFilePath, sigPath string) (*sync.Mutex, int, []string)
}

type Installer interface {
	InstallPackage(ctx context.Context, logE *logrus.Entry, param *installpackage.ParamInstallPackage) error
}

type Downloader interface {
	ReadCloser(ctx context.Context, logE *logrus.Entry, file *download.File) (io.ReadCloser, int64, error)
}

type Executor interface {
	Exec(ctx context.Context, exePath string, args ...string) (int, error)
}

func New(fs afero.Fs, inst Installer, dl Downloader, exe Executor, rt *runtime.Runtime, param *config.Param) *Verifier {
	return &Verifier{
		fs:      fs,
		inst:    inst,
		dl:      dl,
		exe:     exe,
		rt:      rt,
		rootDir: param.RootDir,
	}
}

func (v *Verifier) AddTools(tools ...Tool) {
	v.tools = append(v.tools, tools...)
}

type Verifier struct {
	fs      afero.Fs
	tools   []Tool
	inst    Installer
	dl      Downloader
	exe     Executor
	rt      *runtime.Runtime
	rootDir string
}

func (v *Verifier) Verify(ctx context.Context, logE *logrus.Entry, pkg *config.Package, bodyFile *download.DownloadedFile) error {
	for _, tool := range v.tools {
		if err := v.verifyTool(ctx, logE, tool, pkg, bodyFile); err != nil {
			return logerr.WithFields(err, logrus.Fields{
				"verifier": tool.Package().Package.Name,
			})
		}
	}
	return nil
}

func (v *Verifier) verifyTool(ctx context.Context, logE *logrus.Entry, tool Tool, pkg *config.Package, bodyFile *download.DownloadedFile) error { //nolint:funlen
	// Check if the installed package enables the verifier
	if !tool.Enabled(pkg.PackageInfo) {
		return nil
	}

	tp := tool.Package()
	env := v.rt.Env()
	chksum := tool.Checksums()[env]

	// Check if the runtime (GOOS, GOARCH) supports the verifier.
	if !tp.PackageInfo.CheckSupportedEnvs(v.rt.GOOS, v.rt.GOARCH, v.rt.Env()) {
		return nil
	}

	// Check if the configuration (configuration file, environment variable, command line option) supports the verifier.
	if !tool.SupportedConfig() {
		return nil
	}

	// Install the verification tool
	if err := v.inst.InstallPackage(ctx, logE, &installpackage.ParamInstallPackage{
		Checksums: checksum.New(), // Check minisign's checksum but not update aqua-checksums.json
		Pkg:       tp,
		Checksum: &checksum.Checksum{
			Algorithm: "sha256",
			Checksum:  chksum,
		},
		DisablePolicy: true,
	}); err != nil {
		return err
	}

	// Download the signature
	sig, sigPath, err := tool.Signature(ctx, logE)
	if err != nil {
		return fmt.Errorf("get the signature: %w", err)
	}
	if sig != nil {
		s, err := v.downloadSignature(ctx, logE, tp, sig, sigPath)
		if err != nil {
			return err
		}
		defer v.fs.Remove(s) //nolint:errcheck
		sigPath = s
	}

	tp.PackageInfo.OverrideByRuntime(v.rt)
	exePath, err := pkg.ExePath(v.rootDir, pkg.PackageInfo.GetFiles()[0], v.rt)
	if err != nil {
		return fmt.Errorf("get an executable file path of minisign: %w", err)
	}

	verifiedFilePath, err := bodyFile.Path()
	if err != nil {
		return fmt.Errorf("get a temporary file path: %w", err)
	}

	// Execute the verification tool
	mutex, retryCount, args := tool.Command(verifiedFilePath, sigPath)
	for i := range retryCount {
		mutex.Lock()
		defer mutex.Unlock()
		_, err := v.exe.Exec(ctx, exePath, args...)
		if err == nil {
			return nil
		}
		if i == 4 { //nolint:mnd
			break
		}
		if err := wait(ctx, logE, i+1); err != nil {
			return err
		}
	}
	return errors.New("verify the package")
}

func (v *Verifier) getSignatureReader(ctx context.Context, logE *logrus.Entry, tool *config.Package, sig *registry.DownloadedFile) (io.ReadCloser, error) {
	f, err := download.ConvertDownloadedFileToFile(sig, &download.File{
		RepoOwner: tool.PackageInfo.RepoOwner,
		RepoName:  tool.PackageInfo.RepoName,
		Version:   tool.Package.Version,
	}, v.rt, tool.TemplateArtifact(v.rt, tool.PackageInfo.Asset))
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	rc, _, err := v.dl.ReadCloser(ctx, logE, f)
	if err != nil {
		return nil, fmt.Errorf("download a signature: %w", err)
	}
	return rc, err
}

func (v *Verifier) getSignatureWriter(sigPath string) (afero.File, string, error) {
	if sigPath != "" {
		f, err := v.fs.Create(sigPath)
		return f, sigPath, err
	}
	s, err := afero.TempFile(v.fs, "", "")
	return s, s.Name(), err
}

func (v *Verifier) downloadSignature(ctx context.Context, logE *logrus.Entry, tool *config.Package, sig *registry.DownloadedFile, sigPath string) (string, error) {
	rc, err := v.getSignatureReader(ctx, logE, tool, sig)
	if err != nil {
		return "", fmt.Errorf("download a signature: %w", err)
	}
	defer rc.Close()

	signatureFile, sigPath, err := v.getSignatureWriter(sigPath)
	if err != nil {
		return "", err
	}
	defer signatureFile.Close()
	if _, err := io.Copy(signatureFile, rc); err != nil {
		return "", fmt.Errorf("copy a signature to a temporary file: %w", err)
	}
	return sigPath, nil
}

func wait(ctx context.Context, logE *logrus.Entry, retryCount int) error {
	randGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))       //nolint:gosec
	waitTime := time.Duration(randGenerator.Intn(1000)) * time.Millisecond //nolint:mnd
	logE.WithFields(logrus.Fields{
		"retry_count": retryCount,
		"wait_time":   waitTime,
	}).Info("Verification failed temporarily, retring")
	if err := timer.Wait(ctx, waitTime); err != nil {
		return fmt.Errorf("wait verification: %w", err)
	}
	return nil
}
