package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/suzuki-shunsuke/go-timeout/timeout"
)

func (ctrl *Controller) Exec(ctx context.Context, param *Param, args []string) error {
	if len(args) == 0 {
		return errors.New("command is required")
	}
	exeName := args[0]
	cfg := &Config{}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	if err := ctrl.readConfig(wd, param.ConfigFilePath, cfg); err != nil {
		return err
	}
	inlineRepo := make(map[string]*PackageInfo, len(cfg.InlineRepository))
	for _, pkgInfo := range cfg.InlineRepository {
		inlineRepo[pkgInfo.Name] = pkgInfo
	}
	for _, pkg := range cfg.Packages {
		if pkg.Name != exeName {
			continue
		}

		if err := ctrl.installPackage(ctx, inlineRepo, pkg, cfg); err != nil {
			return err
		}

		pkgInfo, ok := inlineRepo[pkg.Name]
		if !ok {
			return fmt.Errorf("repository isn't found %s", pkg.Name)
		}

		return ctrl.exec(ctx, exeName, pkg, pkgInfo, args[1:])
	}
	return nil
}

func (ctrl *Controller) exec(ctx context.Context, exeName string, pkg *Package, pkgInfo *PackageInfo, args []string) error {
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
	exePath := filepath.Join(pkgPath, exeName) // TODO
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
