package outputshell

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func (c *Controller) OutputShell(ctx context.Context, logE *logrus.Entry, param *config.Param) error { //nolint:funlen
	shellPath := filepath.Join(c.rootDir, "shell", strconv.Itoa(param.Ppid), "shell.json")

	oldPaths, err := c.readOldPaths(shellPath)
	if err != nil {
		return err
	}

	shell := &Shell{
		Env: &Env{
			Path: &Path{
				Values: []string{},
			},
		},
	}
	for _, cfgFilePath := range c.configFinder.Finds(param.PWD, param.ConfigFilePath) {
		if err := c.handleConfig(ctx, logE, cfgFilePath, param, shell); err != nil {
			return err
		}
	}

	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		if f, err := afero.Exists(c.fs, cfgFilePath); err != nil {
			return fmt.Errorf("check if a global configuration file exists: %w", err)
		} else if !f {
			continue
		}
		if err := c.handleConfig(ctx, logE, cfgFilePath, param, shell); err != nil {
			return err
		}
	}

	newPS, updated := c.getNewPS(param, shell, oldPaths)

	if updated {
		fmt.Fprintln(c.stdout, "export PATH="+strings.Join(newPS, param.PathListSeparator))
	}

	if err := c.saveShell(shellPath, shell); err != nil {
		return err
	}

	return nil
}

func (c *Controller) readOldPaths(shellPath string) (map[string]struct{}, error) {
	oldShell := &Shell{}
	if err := c.readShell(shellPath, oldShell); err != nil {
		return nil, err
	}

	oldPaths := make(map[string]struct{}, len(oldShell.GetPaths()))
	for _, p := range oldShell.GetPaths() {
		oldPaths[p] = struct{}{}
	}
	return oldPaths, nil
}

func (c *Controller) getNewPS(param *config.Param, shell *Shell, oldPaths map[string]struct{}) ([]string, bool) {
	paths := make(map[string]struct{}, len(shell.GetPaths()))
	for _, p := range shell.GetPaths() {
		paths[p] = struct{}{}
	}

	ps := strings.Split(param.EnvPath, param.PathListSeparator)
	psMap := make(map[string]struct{}, len(ps))
	for _, p := range ps {
		psMap[p] = struct{}{}
	}

	newPS := make([]string, 0, len(ps))

	updated := false
	for k := range paths {
		if _, ok := psMap[k]; ok {
			continue
		}
		// Add path to head
		newPS = append(newPS, k)
		updated = true
	}

	for _, p := range ps {
		if _, ok := paths[p]; ok {
			newPS = append(newPS, p)
			continue
		}
		if _, ok := oldPaths[p]; ok {
			// Remove path
			updated = true
			continue
		}
		newPS = append(newPS, p)
	}
	return newPS, updated
}

func (c *Controller) saveShell(shellPath string, shell *Shell) error {
	if err := osfile.MkdirAll(c.fs, filepath.Dir(shellPath)); err != nil {
		return fmt.Errorf("create a directory to store shell.json: %w", err)
	}
	f, err := c.fs.Create(shellPath)
	if err != nil {
		return fmt.Errorf("create a shell.json: %w", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(shell); err != nil {
		return fmt.Errorf("write shell.json: %w", err)
	}
	return nil
}

func (c *Controller) readShell(shellPath string, shell *Shell) error {
	if f, err := afero.Exists(c.fs, shellPath); err != nil {
		return fmt.Errorf("check if shell.json exists: %w", err)
	} else if !f {
		return nil
	}
	f, err := c.fs.Open(shellPath)
	if err != nil {
		return fmt.Errorf("open a shell.json: %w", err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(shell); err != nil {
		return fmt.Errorf("read a shell.json: %w", err)
	}
	return nil
}

func (c *Controller) handleConfig(ctx context.Context, logE *logrus.Entry, cfgFilePath string, param *config.Param, shell *Shell) error {
	cfg := &aqua.Config{}
	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	if err := c.configReader.Read(logE, cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	var checksums *checksum.Checksums
	if cfg.ChecksumEnabled(param.EnforceChecksum, param.Checksum) {
		checksums = checksum.New()
		checksumFilePath, err := checksum.GetChecksumFilePathFromConfigFilePath(c.fs, cfgFilePath)
		if err != nil {
			return err //nolint:wrapcheck
		}
		if err := checksums.ReadFile(c.fs, checksumFilePath); err != nil {
			return fmt.Errorf("read a checksum JSON: %w", err)
		}
		defer func() {
			if err := checksums.UpdateFile(c.fs, checksumFilePath); err != nil {
				logE.WithError(err).Error("update a checksum file")
			}
		}()
	}

	registryContents, err := c.registryInstaller.InstallRegistries(ctx, logE, cfg, cfgFilePath, checksums)
	if err != nil {
		return err //nolint:wrapcheck
	}
	pkgs, failed := config.ListPackages(logE, cfg, c.runtime, registryContents)
	if failed {
		return errors.New("failed to list packages")
	}
	for _, pkg := range pkgs {
		if err := c.handlePkg(shell, pkg); err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) handlePkg(shell *Shell, pkg *config.Package) error {
	if pkg.PackageInfo.Shell == nil {
		return nil
	}
	p, ok := pkg.PackageInfo.Shell.Env["PATH"]
	if !ok {
		return nil
	}
	newP, err := pkg.RenderTemplateString(p, c.runtime)
	if err != nil {
		return fmt.Errorf("render added $PATH: %w", err)
	}
	pkgPath, err := pkg.PkgPath(c.rootDir, c.runtime)
	if err != nil {
		return fmt.Errorf("get the installed package path: %w", err)
	}
	shell.Env.Path.Values = append(shell.Env.Path.Values, filepath.Join(pkgPath, filepath.FromSlash(newP)))
	return nil
}
