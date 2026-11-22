package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"go.yaml.in/yaml/v2"
)

var errDownloadOnlyGitHubContent = errors.New("only a github_content registry can be downloaded")

// DownloadRegistry downloads a github_content registry whatever is installed, checks it
// against checksums the way InstallRegistry does (recording a checksum the set doesn't
// have), and installs it in place of what was there.
//
// It is for recording a registry's checksum, which has to be the checksum of what the
// registry serves: a file already on disk may be stale or edited. The files are written
// to temporary files and renamed into place, the JSON made from a YAML registry included,
// so another aqua reading the same registry never sees a file half written, and a failed
// download leaves the installed files as they were.
func (is *Installer) DownloadRegistry(ctx context.Context, logger *slog.Logger, regist *aqua.Registry, cfgFilePath string, checksums *checksum.Checksums) (*registry.Config, error) {
	if err := regist.Validate(); err != nil {
		return nil, fmt.Errorf("validate the registry: %w", err)
	}
	if regist.Type != aqua.RegistryTypeGitHubContent {
		return nil, errDownloadOnlyGitHubContent
	}
	registryFilePath, err := regist.FilePath(is.param.RootDir, cfgFilePath)
	if err != nil {
		return nil, fmt.Errorf("get a registry file path: %w", err)
	}

	content, err := is.downloadGitHubContent(ctx, logger, regist)
	if err != nil {
		return nil, err
	}
	if err := checksum.CheckRegistry(regist, checksums, content); err != nil {
		return nil, fmt.Errorf("check a registry's checksum: %w", err)
	}

	registryContent, err := parseRegistry(registryFilePath, content)
	if err != nil {
		return nil, err
	}
	if err := writeRegistry(registryFilePath, content, registryContent); err != nil {
		return nil, err
	}
	return registryContent, nil
}

func parseRegistry(registryFilePath string, content []byte) (*registry.Config, error) {
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

// writeRegistry installs a downloaded registry file in place of what was there.
func writeRegistry(registryFilePath string, content []byte, registryContent *registry.Config) error {
	if err := osfile.MkdirAll(filepath.Dir(registryFilePath)); err != nil {
		return fmt.Errorf("create the parent directory of the registry file: %w", err)
	}
	if !isJSON(registryFilePath) {
		// InstallRegistry reads the JSON made from a YAML registry before the YAML, so
		// it is replaced as well. It is written first: a JSON left from the previous
		// file beside the new YAML would be what is read.
		b, err := json.Marshal(registryContent)
		if err != nil {
			return fmt.Errorf("encode a registry as JSON: %w", err)
		}
		if err := writeFileAtomically(registryFilePath+jsonSuffix, b); err != nil {
			return err
		}
	}
	return writeFileAtomically(registryFilePath, content)
}

func (is *Installer) downloadGitHubContent(ctx context.Context, logger *slog.Logger, regist *aqua.Registry) ([]byte, error) {
	file, err := is.registryDownloader.DownloadGitHubContentFile(ctx, logger, &domain.GitHubContentFileParam{
		RepoOwner: regist.RepoOwner,
		RepoName:  regist.RepoName,
		Ref:       regist.Ref,
		Path:      regist.Path,
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	defer file.Close()
	content, err := file.Byte()
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return content, nil
}

// writeFileAtomically writes b to a temporary file beside path and renames it to path.
func writeFileAtomically(path string, b []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create a temporary file: %w", err)
	}
	tmp := f.Name()
	if _, err := f.Write(b); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("write a temporary file: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("close a temporary file: %w", err)
	}
	if err := os.Chmod(tmp, registryFilePermission); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("change the mode of a temporary file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("rename a temporary file into place: %w", err)
	}
	return nil
}
