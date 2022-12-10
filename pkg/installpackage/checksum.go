package installpackage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/cosign"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (inst *Installer) extractChecksum(pkg *config.Package, assetName string, checksumFile []byte) (string, error) {
	pkgInfo := pkg.PackageInfo

	if pkgInfo.Checksum.FileFormat == "raw" {
		return strings.TrimSpace(string(checksumFile)), nil
	}

	m, s, err := inst.checksumFileParser.ParseChecksumFile(string(checksumFile), pkg)
	if err != nil {
		return "", fmt.Errorf("parse a checksum file: %w", err)
	}
	if s != "" {
		return s, nil
	}

	return m[assetName], nil
}

func (inst *Installer) verifyChecksumFileWithCosign(ctx context.Context, logE *logrus.Entry, pkg *config.Package, cos *registry.Cosign, b []byte) error {
	if !inst.cosign.HasCosign() {
		logE.Info("skip verifying a signature of checksum file with Cosign, because Cosign isn't inatalled")
		return nil
	}
	f, err := afero.TempFile(inst.fs, "", "")
	if err != nil {
		return fmt.Errorf("create a temporal file: %w", err)
	}
	if _, err := f.Write(b); err != nil {
		return fmt.Errorf("write contents to a temporal file: %w", err)
	}
	defer inst.fs.Remove(f.Name()) //nolint:errcheck
	c, err := pkg.RenderCosign(cos, inst.runtime)
	if err != nil {
		return fmt.Errorf("render cosign options: %w", err)
	}

	if c.Signature != nil {
		sigFile, err := afero.TempFile(inst.fs, "", "")
		if err != nil {
			return fmt.Errorf("create a temporal file: %w", err)
		}
		defer inst.fs.Remove(sigFile.Name())
		if err := inst.downloadCosignFile(ctx, logE, pkg, c, c.Signature, sigFile); err != nil {
			return fmt.Errorf("download a signature: %w", err)
		}
		c.Opts = append(c.Opts, "--signature", sigFile.Name())
	}
	if c.Key != nil {
		keyFile, err := afero.TempFile(inst.fs, "", "")
		if err != nil {
			return fmt.Errorf("create a temporal file: %w", err)
		}
		defer inst.fs.Remove(keyFile.Name())
		if err := inst.downloadCosignFile(ctx, logE, pkg, c, c.Signature, keyFile); err != nil {
			return fmt.Errorf("download a key: %w", err)
		}
		c.Opts = append(c.Opts, "--key", keyFile.Name())
	}
	if c.Certificate != nil {
		certFile, err := afero.TempFile(inst.fs, "", "")
		if err != nil {
			return fmt.Errorf("create a temporal file: %w", err)
		}
		defer inst.fs.Remove(certFile.Name())
		if err := inst.downloadCosignFile(ctx, logE, pkg, c, c.Signature, certFile); err != nil {
			return fmt.Errorf("download a certificate: %w", err)
		}
		c.Opts = append(c.Opts, "--certificate", certFile.Name())
	}

	logE.Info("verify a checksum file with Cosign")
	if err := inst.cosign.Verify(ctx, &cosign.ParamVerify{
		Opts:               c.Opts,
		CosignExperimental: c.CosignExperimental,
		Target:             f.Name(),
	}); err != nil {
		return fmt.Errorf("verify a checksum file with Cosign: %w", err)
	}
	return nil
}

func convertDownloadedFileToFile(file *registry.DownloadedFile, pkg *config.Package, rt *runtime.Runtime) (*download.File, error) {
	f := &download.File{
		Type:      file.Type,
		RepoOwner: file.RepoOwner,
		RepoName:  file.RepoName,
		Version:   pkg.Package.Version,
	}
	pkgInfo := pkg.PackageInfo
	switch file.Type {
	case "github_release":
		if f.RepoOwner == "" {
			f.RepoOwner = pkgInfo.RepoOwner
		}
		if f.RepoName == "" {
			f.RepoName = pkgInfo.RepoName
		}
		if file.Asset == nil {
			return nil, errors.New("asset is required")
		}
		asset, err := pkg.RenderTemplateString(*file.Asset, rt)
		if err != nil {
			return nil, err
		}
		f.Asset = asset
		return f, nil
	case "http":
		if file.URL == nil {
			return nil, errors.New("url is required")
		}
		u, err := pkg.RenderTemplateString(*file.URL, rt)
		if err != nil {
			return nil, err
		}
		f.URL = u
		return f, nil
	}
	return nil, logerr.WithFields(errors.New("invalid file type"), logrus.Fields{
		"file_type": file.Type,
	})
}

func (inst *Installer) downloadCosignFile(ctx context.Context, logE *logrus.Entry, pkg *config.Package, cos *registry.Cosign, file *registry.DownloadedFile, tf io.Writer) error {
	f, err := convertDownloadedFileToFile(file, pkg, inst.runtime)
	if err != nil {
		return err
	}
	rc, _, err := inst.downloader.GetReadCloser(ctx, f, logE)
	if err != nil {
		return err
	}
	if _, err := io.Copy(tf, rc); err != nil {
		return err
	}
	return nil
}

func (inst *Installer) dlAndExtractChecksum(ctx context.Context, logE *logrus.Entry, pkg *config.Package, assetName string) (string, error) {
	file, _, err := inst.checksumDownloader.DownloadChecksum(ctx, logE, inst.runtime, pkg)
	if err != nil {
		return "", fmt.Errorf("download a checksum file: %w", err)
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("read a checksum file: %w", err)
	}

	if cos := pkg.PackageInfo.Checksum.GetCosign(); cos.GetEnabled() {
		if err := inst.verifyChecksumFileWithCosign(ctx, logE, pkg, cos, b); err != nil {
			return "", fmt.Errorf("verify a checksum file with Cosign: %w", err)
		}
	}

	c, err := inst.extractChecksum(pkg, assetName, b)
	if err != nil {
		return "", err
	}
	if c == "" {
		return "", errors.New("checksum isn't found in a checksum file")
	}
	return c, nil
}

type ParamVerifyChecksum struct {
	ChecksumID string
	Checksum   *checksum.Checksum
	Checksums  *checksum.Checksums
	Pkg        *config.Package
	AssetName  string
	Body       io.Reader
	TempDir    string
}

func copyAsset(fs afero.Fs, tempFilePath string, body io.Reader) error {
	file, err := fs.Create(tempFilePath)
	if err != nil {
		return fmt.Errorf("create a temporal file: %w", logerr.WithFields(err, logrus.Fields{
			"temp_file": tempFilePath,
		}))
	}
	defer file.Close()

	if _, err := io.Copy(file, body); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}

func (inst *Installer) verifyChecksum(ctx context.Context, logE *logrus.Entry, param *ParamVerifyChecksum) (io.ReadCloser, error) { //nolint:cyclop,funlen,gocognit
	pkg := param.Pkg
	pkgInfo := pkg.PackageInfo
	checksums := param.Checksums
	chksum := param.Checksum
	checksumID := param.ChecksumID
	tempDir := param.TempDir

	// Download an asset in a temporal directory
	// Calculate the checksum of download asset
	// Download a checksum file
	// Extract the checksum from the checksum file
	// Compare the checksum
	// Store the checksum to aqua-checksums.json

	assetName := param.AssetName
	// If pkgInfo.Type is "github_archive", AssetName is empty.
	// filepath.Base("") returns "."
	if assetName != "" {
		// For github_content
		assetName = filepath.Base(assetName)
	}
	tempFilePath := filepath.Join(tempDir, assetName)
	if assetName == "" && (pkgInfo.Type == "github_archive" || pkgInfo.Type == "go") {
		tempFilePath = filepath.Join(tempDir, "archive.tar.gz")
	}
	if err := copyAsset(inst.fs, tempFilePath, param.Body); err != nil {
		return nil, err
	}

	if chksum == nil && pkgInfo.Checksum.GetEnabled() {
		logE.Info("downloading a checksum file")
		c, err := inst.dlAndExtractChecksum(ctx, logE, pkg, assetName)
		if err != nil {
			return nil, logerr.WithFields(err, logrus.Fields{ //nolint:wrapcheck
				"asset_name": assetName,
			})
		}
		chksum = &checksum.Checksum{
			ID:        checksumID,
			Checksum:  c,
			Algorithm: pkgInfo.Checksum.GetAlgorithm(),
		}
		checksums.Set(checksumID, chksum)
	}

	if chksum != nil {
		chksum.Checksum = strings.ToUpper(chksum.Checksum)
	}

	algorithm := "sha512"
	if chksum != nil {
		algorithm = chksum.Algorithm
	}
	calculatedSum, err := inst.checksumCalculator.Calculate(inst.fs, tempFilePath, algorithm)
	if err != nil {
		return nil, fmt.Errorf("calculate a checksum of downloaded file: %w", logerr.WithFields(err, logrus.Fields{
			"temp_file": tempFilePath,
		}))
	}
	calculatedSum = strings.ToUpper(calculatedSum)

	if chksum != nil && calculatedSum != chksum.Checksum {
		return nil, logerr.WithFields(errInvalidChecksum, logrus.Fields{ //nolint:wrapcheck
			"actual_checksum":   calculatedSum,
			"expected_checksum": chksum.Checksum,
		})
	}

	if chksum == nil {
		logE.WithFields(logrus.Fields{
			"checksum_id": checksumID,
			"checksum":    calculatedSum,
		}).Debug("set a calculated checksum")
		chksum = &checksum.Checksum{
			ID:        checksumID,
			Checksum:  calculatedSum,
			Algorithm: pkgInfo.Checksum.GetAlgorithm(),
		}
	}
	checksums.Set(checksumID, chksum)

	// Verify with Cosign
	if cos := pkg.PackageInfo.Cosign; cos != nil { //nolint:nestif
		if inst.cosign.HasCosign() {
			c, err := pkg.RenderCosign(cos, inst.runtime)
			if err != nil {
				return nil, fmt.Errorf("render cosign options: %w", err)
			}
			if err := inst.cosign.Verify(ctx, &cosign.ParamVerify{
				CosignExperimental: c.CosignExperimental,
				Opts:               c.Opts,
				Target:             tempFilePath,
			}); err != nil {
				return nil, fmt.Errorf("verify with Cosign: %w", err)
			}
		}
	}

	readFile, err := inst.fs.Open(tempFilePath)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return readFile, nil
}
