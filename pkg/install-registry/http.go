package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"go.yaml.in/yaml/v2"
)

func (is *Installer) getHTTPRegistry(ctx context.Context, logE *logrus.Entry, regist *aqua.Registry, registryFilePath string, checksums *checksum.Checksums) (*registry.Config, error) {
	// Render the URL with the version
	renderedURL, err := regist.RenderURL()
	if err != nil {
		return nil, fmt.Errorf("render registry URL: %w", err)
	}

	logE.WithFields(logrus.Fields{
		"registry_url": renderedURL,
		"version":      regist.Version,
	}).Debug("downloading HTTP registry")

	// Download the registry file
	body, _, err := is.httpDownloader.Download(ctx, renderedURL)
	if err != nil {
		return nil, logerr.WithFields(err, logrus.Fields{ //nolint:wrapcheck
			"registry_url": renderedURL,
		})
	}
	defer body.Close()

	// Read the content
	content, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read HTTP registry content: %w", err)
	}

	// Verify checksum if provided
	if checksums != nil {
		if err := checksum.CheckRegistry(regist, checksums, content); err != nil {
			return nil, fmt.Errorf("check a registry's checksum: %w", err)
		}
	}

	// Create the parent directory
	if err := osfile.MkdirAll(is.fs, filepath.Dir(registryFilePath)); err != nil {
		return nil, fmt.Errorf("create the parent directory of the registry file: %w", err)
	}

	// Handle different formats
	format := regist.Format
	if format == "" {
		format = "raw"
	}

	switch format {
	case "raw":
		return is.handleRawHTTPRegistry(registryFilePath, content)
	case "tar.gz":
		return is.handleTarGzHTTPRegistry(ctx, logE, registryFilePath, content, regist)
	default:
		return nil, logerr.WithFields(errInvalidRegistryFormat, logrus.Fields{ //nolint:wrapcheck
			"format": format,
		})
	}
}

func (is *Installer) handleRawHTTPRegistry(registryFilePath string, content []byte) (*registry.Config, error) {
	// Write the content to the registry file
	if err := afero.WriteFile(is.fs, registryFilePath, content, registryFilePermission); err != nil {
		return nil, fmt.Errorf("write the registry file: %w", err)
	}

	// Parse the registry content
	registryContent := &registry.Config{}
	if isJSON(registryFilePath) {
		if err := json.Unmarshal(content, registryContent); err != nil {
			return nil, fmt.Errorf("parse the registry configuration file as JSON: %w", err)
		}
		return registryContent, nil
	}
	if err := yaml.Unmarshal(content, registryContent); err != nil {
		return nil, fmt.Errorf("parse the registry configuration file as YAML: %w", err)
	}
	return registryContent, nil
}

func (is *Installer) handleTarGzHTTPRegistry(ctx context.Context, logE *logrus.Entry, registryFilePath string, content []byte, regist *aqua.Registry) (*registry.Config, error) {
	// Create a temporary file for the tar.gz content
	tempDir := filepath.Join(filepath.Dir(registryFilePath), ".tmp")
	if err := osfile.MkdirAll(is.fs, tempDir); err != nil {
		return nil, fmt.Errorf("create temp directory: %w", err)
	}
	defer func() {
		if err := is.fs.RemoveAll(tempDir); err != nil {
			logE.WithError(err).Warn("failed to remove temp directory")
		}
	}()

	tempArchivePath := filepath.Join(tempDir, "registry.tar.gz")
	if err := afero.WriteFile(is.fs, tempArchivePath, content, registryFilePermission); err != nil {
		return nil, fmt.Errorf("write temp archive file: %w", err)
	}

	// Extract the archive
	extractDir := filepath.Join(tempDir, "extracted")
	if err := osfile.MkdirAll(is.fs, extractDir); err != nil {
		return nil, fmt.Errorf("create extraction directory: %w", err)
	}

	// Use unarchive package to extract
	unarchiver := unarchive.New(nil, is.fs)
	src := &unarchive.File{
		Body:     &tempFileBody{path: tempArchivePath, fs: is.fs},
		Filename: "registry.tar.gz",
		Type:     "tar.gz",
	}
	if err := unarchiver.Unarchive(ctx, logE, src, extractDir); err != nil {
		return nil, fmt.Errorf("extract tar.gz archive: %w", err)
	}

	// Find the registry file in the extracted content
	// Look for registry.yaml or registry.json in the extracted directory
	registryPath := regist.Path
	if registryPath == "" {
		// Try to find registry.yaml or registry.json
		if _, err := is.fs.Stat(filepath.Join(extractDir, "registry.yaml")); err == nil {
			registryPath = "registry.yaml"
		} else if _, err := is.fs.Stat(filepath.Join(extractDir, "registry.json")); err == nil {
			registryPath = "registry.json"
		} else {
			return nil, errRegistryFileNotFoundInArchive
		}
	}

	extractedRegistryPath := filepath.Join(extractDir, registryPath)
	extractedContent, err := afero.ReadFile(is.fs, extractedRegistryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, logerr.WithFields(errRegistryFileNotFoundInArchive, logrus.Fields{ //nolint:wrapcheck
				"path": registryPath,
			})
		}
		return nil, fmt.Errorf("read extracted registry file: %w", err)
	}

	// Write the extracted registry content to the final location
	// Determine the final file path based on the extracted file type
	finalPath := registryFilePath
	if isJSON(extractedRegistryPath) && !isJSON(registryFilePath) {
		finalPath = registryFilePath + jsonSuffix
	}

	if err := afero.WriteFile(is.fs, finalPath, extractedContent, registryFilePermission); err != nil {
		return nil, fmt.Errorf("write the registry file: %w", err)
	}

	// Parse the registry content
	registryContent := &registry.Config{}
	if isJSON(extractedRegistryPath) {
		if err := json.Unmarshal(extractedContent, registryContent); err != nil {
			return nil, fmt.Errorf("parse the registry configuration file as JSON: %w", err)
		}
		return registryContent, nil
	}
	if err := yaml.Unmarshal(extractedContent, registryContent); err != nil {
		return nil, fmt.Errorf("parse the registry configuration file as YAML: %w", err)
	}
	return registryContent, nil
}

// tempFileBody implements the unarchive.DownloadedFile interface for a local file.
type tempFileBody struct {
	path string
	fs   afero.Fs
}

func (t *tempFileBody) Path() (string, error) {
	return t.path, nil
}

func (t *tempFileBody) ReadLast() (io.ReadCloser, error) {
	return t.fs.Open(t.path) //nolint:wrapcheck
}

func (t *tempFileBody) Wrap(w io.Writer) io.Writer {
	return w
}
