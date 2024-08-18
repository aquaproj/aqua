package setshell

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

func (c *Controller) SetShell(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	shellPath := filepath.Join(c.rootDir, "shell", strconv.Itoa(param.Ppid), "shell.json")

	oldShell := &Shell{}
	if err := c.readShell(shellPath, oldShell); err != nil {
		return err
	}

	oldPaths := make(map[string]struct{}, len(oldShell.GetPaths()))
	for _, p := range oldShell.GetPaths() {
		oldPaths[p] = struct{}{}
	}

	shell := &Shell{}
	for _, cfgFilePath := range c.configFinder.Finds(param.PWD, param.ConfigFilePath) {
		if err := c.handleConfig(ctx, logE, cfgFilePath, param, shell); err != nil {
			return err
		}
	}

	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		if err := c.handleConfig(ctx, logE, cfgFilePath, param, shell); err != nil {
			return err
		}
	}

	paths := make(map[string]struct{}, len(shell.GetPaths()))
	for _, p := range shell.GetPaths() {
		paths[p] = struct{}{}
	}

	ps := strings.Split(param.EnvPath, string(param.PathListSeparator))
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

	if updated {
		fmt.Fprintf(c.stdout, "export PATH=%s\n", strings.Join(newPS, param.PathListSeparator))
	}

	if err := c.saveShell(shellPath, shell); err != nil {
		return err
	}

	return nil
}

func (c *Controller) saveShell(shellPath string, shell *Shell) error {
	if err := osfile.MkdirAll(c.fs, filepath.Dir(shellPath)); err != nil {
		return err
	}
	f, err := c.fs.Create(shellPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(shell); err != nil {
		return err
	}
	return nil
}

func (c *Controller) readShell(shellPath string, shell *Shell) error {
	if f, err := afero.Exists(c.fs, shellPath); err != nil {
		return err
	} else if !f {
		return nil
	}
	f, err := c.fs.Open(shellPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(shell); err != nil {
		return err
	}
	return nil
}

func (c *Controller) handleConfig(ctx context.Context, logE *logrus.Entry, cfgFilePath string, param *config.Param, oldShell *Shell) error {
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
		if pkg.PackageInfo.Shell != nil {
			p, ok := pkg.PackageInfo.Shell.Env["PATH"]
			if !ok {
				continue
			}
			newP, err := pkg.RenderTemplateString(p, c.runtime)
			if err != nil {
				return err
			}
			pkgPath, err := pkg.PkgPath(c.rootDir, c.runtime)
			if err != nil {
				return err
			}
			oldShell.Env.Path.Values = append(oldShell.Env.Path.Values, filepath.Join(pkgPath, filepath.FromSlash(newP)))
		}
	}

	return nil
}
