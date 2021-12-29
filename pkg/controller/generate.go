package controller

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"gopkg.in/yaml.v2"
)

type FindingPackage struct {
	PackageInfo  *MergedPackageInfo
	RegistryName string
}

func (ctrl *Controller) Generate(ctx context.Context, param *Param, args ...string) error { //nolint:cyclop,funlen
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	cfgFilePath := ctrl.getConfigFilePath(wd, param.ConfigFilePath)
	if cfgFilePath == "" {
		return errConfigFileNotFound
	}
	cfg := &Config{}
	if err := ctrl.readConfig(cfgFilePath, cfg); err != nil {
		return err
	}
	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("configuration is invalid: %w", err)
	}
	registryContents, err := ctrl.installRegistries(ctx, cfg, cfgFilePath)
	if err != nil {
		return err
	}

	if param.File != "" || len(args) != 0 {
		return ctrl.outputListedPkgs(ctx, param, registryContents, args...)
	}

	var pkgs []*FindingPackage
	for registryName, registryContent := range registryContents {
		for _, pkg := range registryContent.PackageInfos {
			pkgs = append(pkgs, &FindingPackage{
				PackageInfo:  pkg,
				RegistryName: registryName,
			})
		}
	}
	idx, err := fuzzyfinder.Find(pkgs, func(i int) string {
		pkg := pkgs[i]
		files := pkg.PackageInfo.GetFiles()
		fileNames := make([]string, len(files))
		for i, file := range files {
			fileNames[i] = file.Name
		}
		fileNamesStr := strings.Join(fileNames, ", ")
		pkgName := pkg.PackageInfo.GetName()
		if strings.HasSuffix(pkgName, "/"+fileNamesStr) || pkgName == fileNamesStr {
			return fmt.Sprintf("%s (%s)", pkgName, pkg.RegistryName)
		}
		return fmt.Sprintf("%s (%s) (%s)", pkgName, pkg.RegistryName, fileNamesStr)
	},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i < 0 {
				return ""
			}
			pkg := pkgs[i]
			return fmt.Sprintf("%s\n\n%s\n%s",
				pkg.PackageInfo.GetName(),
				pkg.PackageInfo.GetLink(),
				formatDescription(pkg.PackageInfo.GetDescription(), w/2-8)) //nolint:gomnd
		}))
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil
		}
		return fmt.Errorf("find the package: %w", err)
	}
	pkg := pkgs[idx]
	outputPkg := ctrl.getOutputtedPkg(ctx, pkg)
	if err := yaml.NewEncoder(ctrl.Stdout).Encode([]interface{}{outputPkg}); err != nil {
		return fmt.Errorf("output generated package configuration: %w", err)
	}
	return nil
}

func formatDescription(desc string, w int) string {
	descRune := []rune(desc)
	lenDescRune := len(descRune)
	lineWidth := w - len([]rune("\n"))
	numOfLines := (lenDescRune / lineWidth) + 1
	descArr := make([]string, numOfLines)
	for i := 0; i < numOfLines; i++ {
		start := i * lineWidth
		end := start + lineWidth
		if i == numOfLines-1 {
			end = lenDescRune
		}
		descArr[i] = string(descRune[start:end])
	}
	return strings.Join(descArr, "\n")
}

func getGeneratePkg(s string) string {
	if !strings.Contains(s, ",") {
		return "standard," + s
	}
	return s
}

func (ctrl *Controller) outputListedPkgs(ctx context.Context, param *Param, registryContents map[string]*RegistryContent, pkgNames ...string) error { //nolint:cyclop
	m := map[string]*FindingPackage{}
	for registryName, registryContent := range registryContents {
		for _, pkg := range registryContent.PackageInfos {
			m[registryName+","+pkg.GetName()] = &FindingPackage{
				PackageInfo:  pkg,
				RegistryName: registryName,
			}
		}
	}

	outputPkgs := []*Package{}
	for _, pkgName := range pkgNames {
		pkgName = getGeneratePkg(pkgName)
		findingPkg, ok := m[pkgName]
		if !ok {
			return logerr.WithFields(errUnknownPkg, logrus.Fields{"package_name": pkgName}) //nolint:wrapcheck
		}
		outputPkg := ctrl.getOutputtedPkg(ctx, findingPkg)
		outputPkgs = append(outputPkgs, outputPkg)
	}

	if param.File != "" { //nolint:nestif
		var file io.Reader
		if param.File == "-" {
			file = ctrl.Stdin
		} else {
			f, err := os.Open(param.File)
			if err != nil {
				return fmt.Errorf("open the package list file: %w", err)
			}
			defer f.Close()
			file = f
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			txt := getGeneratePkg(scanner.Text())
			findingPkg, ok := m[txt]
			if !ok {
				return logerr.WithFields(errUnknownPkg, logrus.Fields{"package_name": txt}) //nolint:wrapcheck
			}
			outputPkg := ctrl.getOutputtedPkg(ctx, findingPkg)
			outputPkgs = append(outputPkgs, outputPkg)
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed to read the file: %w", err)
		}
	}

	if err := yaml.NewEncoder(ctrl.Stdout).Encode(outputPkgs); err != nil {
		return fmt.Errorf("output generated package configuration: %w", err)
	}
	return nil
}

func (ctrl *Controller) getOutputtedGitHubPkg(ctx context.Context, outputPkg *Package, pkgName, repoOwner, repoName string) {
	release, _, err := ctrl.GitHubRepositoryService.GetLatestRelease(ctx, repoOwner, repoName)
	if err != nil {
		logerr.WithError(ctrl.logE(), err).WithFields(logrus.Fields{
			"repo_owner": repoOwner,
			"repo_name":  repoName,
		}).Warn("get the latest release")
		return
	}
	if pkgName == repoOwner+"/"+repoName {
		outputPkg.Name += "@" + release.GetTagName()
		outputPkg.Version = ""
	} else {
		outputPkg.Version = release.GetTagName()
	}
}

func (ctrl *Controller) getOutputtedPkg(ctx context.Context, pkg *FindingPackage) *Package {
	outputPkg := &Package{
		Name:     pkg.PackageInfo.GetName(),
		Registry: pkg.RegistryName,
		Version:  "[SET PACKAGE VERSION]",
	}
	if outputPkg.Registry == "standard" {
		outputPkg.Registry = ""
	}
	if ctrl.GitHubRepositoryService == nil {
		return outputPkg
	}
	pkgInfo := pkg.PackageInfo
	if pkgInfo.HasRepo() {
		ctrl.getOutputtedGitHubPkg(ctx, outputPkg, pkg.PackageInfo.GetName(), pkgInfo.RepoOwner, pkgInfo.RepoName)
		return outputPkg
	}
	return outputPkg
}
