package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
	"github.com/suzuki-shunsuke/go-timeout/timeout"
)

var (
	errCommandIsRequired = errors.New("command is required")
	errCommandIsNotFound = errors.New("command is not found")
)

func (ctrl *Controller) Exec(ctx context.Context, param *Param, args []string) error {
	if len(args) == 0 {
		return errCommandIsRequired
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
	inlineRegistry := make(map[string]*PackageInfo, len(cfg.InlineRegistry))
	for _, pkgInfo := range cfg.InlineRegistry {
		inlineRegistry[pkgInfo.Name] = pkgInfo
	}
	pkg, pkgInfo, file := ctrl.findExecFile(inlineRegistry, cfg, exeName)
	if pkg != nil {
		fileSrc, err := ctrl.getFileSrc(pkg, pkgInfo, file)
		if err != nil {
			return err
		}

		binDir := filepath.Join(filepath.Dir(param.ConfigFilePath), ".aqua", "bin")
		if err := ctrl.installPackage(ctx, inlineRegistry, pkg, binDir, false); err != nil {
			return err
		}

		return ctrl.exec(ctx, pkg, pkgInfo, fileSrc, args[1:])
	}

	return errCommandIsNotFound
}

func (ctrl *Controller) getFileSrc(pkg *Package, pkgInfo *PackageInfo, file *File) (string, error) {
	assetName, err := pkgInfo.RenderAsset(pkg)
	if err != nil {
		return "", fmt.Errorf("render the asset name: %w", err)
	}
	if isUnarchived(pkgInfo.ArchiveType, assetName) {
		return assetName, nil
	}
	if file.Src == nil {
		return file.Name, nil
	}
	src, err := file.RenderSrc(pkg, pkgInfo)
	if err != nil {
		return "", fmt.Errorf("render the template file.src: %w", err)
	}
	return src, nil
}

func (ctrl *Controller) findExecFile(inlineRegistry map[string]*PackageInfo, cfg *Config, exeName string) (*Package, *PackageInfo, *File) {
	for _, pkg := range cfg.Packages {
		pkgInfo, ok := inlineRegistry[pkg.Name]
		if !ok {
			logrus.Warnf("registry isn't found %s", pkg.Name)
			continue
		}
		for _, file := range pkgInfo.Files {
			if file.Name == exeName {
				return pkg, pkgInfo, file
			}
		}
	}
	return nil, nil, nil
}

func isUnarchived(archiveType, assetName string) bool {
	return archiveType == "raw" || (archiveType == "" && filepath.Ext(assetName) == "")
}

func (ctrl *Controller) exec(ctx context.Context, pkg *Package, pkgInfo *PackageInfo, src string, args []string) error {
	assetName, err := pkgInfo.RenderAsset(pkg)
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
		return ecerror.Wrap(err, cmd.ProcessState.ExitCode())
	}
	return nil
}
