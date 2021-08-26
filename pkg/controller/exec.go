package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-timeout/timeout"
)

func (ctrl *Controller) Exec(ctx context.Context, param *Param, args []string) error { //nolint:cyclop
	if len(args) == 0 {
		return errors.New("command is required")
	}
	exeName := args[0]
	cfg := &Config{}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	param.ConfigFilePath = ctrl.getConfigFilePath(wd, param.ConfigFilePath)
	if param.ConfigFilePath == "" {
		return errConfigFileNotFound
	}
	if err := ctrl.readConfig(param.ConfigFilePath, cfg); err != nil {
		return err
	}
	if cfg.BinDir == "" {
		cfg.BinDir = filepath.Join(filepath.Dir(param.ConfigFilePath), ".aqua", "bin")
	}
	inlineRepo := make(map[string]*PackageInfo, len(cfg.InlineRepository))
	for _, pkgInfo := range cfg.InlineRepository {
		inlineRepo[pkgInfo.Name] = pkgInfo
	}
	fileSrc := ""
	for _, pkg := range cfg.Packages {
		pkgInfo, ok := inlineRepo[pkg.Name]
		if !ok {
			logrus.Warnf("repository isn't found %s", pkg.Name)
			continue
		}
		for _, file := range pkgInfo.Files {
			if file.Name == exeName {
				if file.Src == "" {
					fileSrc = file.Name
				} else {
					fileSrc = file.Src
				}
				break
			}
		}
		if fileSrc == "" {
			continue
		}

		if err := ctrl.installPackage(ctx, inlineRepo, pkg, cfg); err != nil {
			return err
		}

		return ctrl.exec(ctx, pkg, pkgInfo, fileSrc, args[1:])
	}
	return errors.New("command is not found")
}

func (ctrl *Controller) exec(ctx context.Context, pkg *Package, pkgInfo *PackageInfo, src string, args []string) error {
	assetName, err := pkgInfo.Artifact.Execute(map[string]interface{}{
		"Package":     pkg,
		"PackageInfo": pkgInfo,
		"OS":          runtime.GOOS,
		"Arch":        runtime.GOARCH,
	})
	if err != nil {
		return fmt.Errorf("render the asset name: %w", err)
	}
	pkgPath := getPkgPath(ctrl.RootDir, pkg, pkgInfo, assetName)
	exePath := filepath.Join(pkgPath, src)
	cmd := exec.Command(exePath, args...)
	cmd.Stdin = ctrl.Stdin
	cmd.Stdout = ctrl.Stdout
	cmd.Stderr = ctrl.Stderr
	runner := timeout.NewRunner(0)
	if err := runner.Run(ctx, cmd); err != nil {
		return fmt.Errorf("execute the command: %w", err)
	}
	return nil
}
