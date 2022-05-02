package generate

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	githubSvc "github.com/aquaproj/aqua/pkg/github"
	registry "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/aquaproj/aqua/pkg/validate"
	constraint "github.com/aquaproj/aqua/pkg/version-constraint"
	"github.com/google/go-github/v39/github"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"gopkg.in/yaml.v2"
)

type Controller struct {
	stdin                   io.Reader
	stdout                  io.Writer
	gitHubRepositoryService githubSvc.RepositoryService
	registryInstaller       registry.Installer
	configFinder            finder.ConfigFinder
	configReader            reader.ConfigReader
}

func New(configFinder finder.ConfigFinder, configReader reader.ConfigReader, registInstaller registry.Installer, gh githubSvc.RepositoryService) *Controller {
	return &Controller{
		stdin:                   os.Stdin,
		stdout:                  os.Stdout,
		configFinder:            configFinder,
		configReader:            configReader,
		registryInstaller:       registInstaller,
		gitHubRepositoryService: gh,
	}
}

// Generate searches packages in registries and outputs the configuration to standard output.
// If no package is specified, the interactive fuzzy finder is launched.
// If the package supports, the latest version is gotten by GitHub API.
func (ctrl *Controller) Generate(ctx context.Context, logE *logrus.Entry, param *config.Param, args ...string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}

	cfgFilePath, err := ctrl.configFinder.Find(wd, param.ConfigFilePath)
	if err != nil {
		return err //nolint:wrapcheck
	}

	list, err := ctrl.generate(ctx, logE, param, cfgFilePath, args...)
	if err != nil {
		return err
	}
	if list == nil {
		return nil
	}
	if !param.Insert {
		if err := yaml.NewEncoder(ctrl.stdout).Encode(list); err != nil {
			return fmt.Errorf("output generated package configuration: %w", err)
		}
		return nil
	}

	return ctrl.generateInsert(cfgFilePath, list)
}

type FindingPackage struct {
	PackageInfo  *config.PackageInfo
	RegistryName string
}

func (ctrl *Controller) generate(ctx context.Context, logE *logrus.Entry, param *config.Param, cfgFilePath string, args ...string) (interface{}, error) { //nolint:cyclop
	cfg := &config.Config{}
	if err := ctrl.configReader.Read(cfgFilePath, cfg); err != nil {
		return nil, err //nolint:wrapcheck
	}
	if err := validate.Config(cfg); err != nil {
		return nil, fmt.Errorf("configuration is invalid: %w", err)
	}
	registryContents, err := ctrl.registryInstaller.InstallRegistries(ctx, cfg, cfgFilePath, logE)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	if param.File != "" || len(args) != 0 {
		return ctrl.outputListedPkgs(ctx, logE, param, registryContents, args...)
	}

	// maps the package and the registry
	var pkgs []*FindingPackage
	for registryName, registryContent := range registryContents {
		for _, pkg := range registryContent.PackageInfos {
			pkgs = append(pkgs, &FindingPackage{
				PackageInfo:  pkg,
				RegistryName: registryName,
			})
		}
	}

	// Launch the fuzzy finder
	idxes, err := ctrl.launchFuzzyFinder(pkgs)
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil, nil //nolint:nilnil
		}
		return nil, fmt.Errorf("find the package: %w", err)
	}
	arr := make([]interface{}, len(idxes))
	for i, idx := range idxes {
		arr[i] = ctrl.getOutputtedPkg(ctx, pkgs[idx], logE)
	}

	return arr, nil
}

func getGeneratePkg(s string) string {
	if !strings.Contains(s, ",") {
		return "standard," + s
	}
	return s
}

func (ctrl *Controller) outputListedPkgs(ctx context.Context, logE *logrus.Entry, param *config.Param, registryContents map[string]*config.RegistryContent, pkgNames ...string) (interface{}, error) {
	m := map[string]*FindingPackage{}
	for registryName, registryContent := range registryContents {
		for _, pkg := range registryContent.PackageInfos {
			m[registryName+","+pkg.GetName()] = &FindingPackage{
				PackageInfo:  pkg,
				RegistryName: registryName,
			}
			for _, alias := range pkg.Aliases {
				m[registryName+","+alias.Name] = &FindingPackage{
					PackageInfo:  pkg,
					RegistryName: registryName,
				}
			}
		}
	}

	outputPkgs := []*config.Package{}
	for _, pkgName := range pkgNames {
		pkgName = getGeneratePkg(pkgName)
		findingPkg, ok := m[pkgName]
		if !ok {
			return nil, logerr.WithFields(errUnknownPkg, logrus.Fields{"package_name": pkgName}) //nolint:wrapcheck
		}
		outputPkg := ctrl.getOutputtedPkg(ctx, findingPkg, logE)
		outputPkgs = append(outputPkgs, outputPkg)
	}

	if param.File != "" {
		pkgs, err := ctrl.readGeneratedPkgsFromFile(ctx, param, outputPkgs, m, logE)
		if err != nil {
			return nil, err
		}
		outputPkgs = pkgs
	}
	return outputPkgs, nil
}

func (ctrl *Controller) readGeneratedPkgsFromFile(ctx context.Context, param *config.Param, outputPkgs []*config.Package, m map[string]*FindingPackage, logE *logrus.Entry) ([]*config.Package, error) {
	var file io.Reader
	if param.File == "-" {
		file = ctrl.stdin
	} else {
		f, err := os.Open(param.File)
		if err != nil {
			return nil, fmt.Errorf("open the package list file: %w", err)
		}
		defer f.Close()
		file = f
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		txt := getGeneratePkg(scanner.Text())
		findingPkg, ok := m[txt]
		if !ok {
			return nil, logerr.WithFields(errUnknownPkg, logrus.Fields{"package_name": txt}) //nolint:wrapcheck
		}
		outputPkg := ctrl.getOutputtedPkg(ctx, findingPkg, logE)
		outputPkgs = append(outputPkgs, outputPkg)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read the file: %w", err)
	}
	return outputPkgs, nil
}

func (ctrl *Controller) listAndGetTagName(ctx context.Context, pkgInfo *config.PackageInfo, logE *logrus.Entry) string {
	repoOwner := pkgInfo.RepoOwner
	repoName := pkgInfo.RepoName
	opt := &github.ListOptions{
		PerPage: 30, //nolint:gomnd
	}
	versionFilter, err := constraint.CompileVersionFilter(*pkgInfo.VersionFilter)
	if err != nil {
		return ""
	}
	for {
		releases, _, err := ctrl.gitHubRepositoryService.ListReleases(ctx, repoOwner, repoName, opt)
		if err != nil {
			logerr.WithError(logE, err).WithFields(logrus.Fields{
				"repo_owner": repoOwner,
				"repo_name":  repoName,
			}).Warn("list releases")
			return ""
		}
		for _, release := range releases {
			if release.GetPrerelease() {
				continue
			}
			f, err := constraint.EvaluateVersionFilter(versionFilter, release.GetTagName())
			if err != nil || !f {
				continue
			}
			return release.GetTagName()
		}
		if len(releases) != opt.PerPage {
			return ""
		}
		opt.Page++
	}
}

func (ctrl *Controller) getOutputtedGitHubPkg(ctx context.Context, outputPkg *config.Package, pkgInfo *config.PackageInfo, logE *logrus.Entry) {
	repoOwner := pkgInfo.RepoOwner
	repoName := pkgInfo.RepoName
	var tagName string
	if pkgInfo.VersionFilter != nil {
		tagName = ctrl.listAndGetTagName(ctx, pkgInfo, logE)
	} else {
		release, _, err := ctrl.gitHubRepositoryService.GetLatestRelease(ctx, repoOwner, repoName)
		if err != nil {
			logerr.WithError(logE, err).WithFields(logrus.Fields{
				"repo_owner": repoOwner,
				"repo_name":  repoName,
			}).Warn("get the latest release")
			return
		}
		tagName = release.GetTagName()
	}
	if pkgName := pkgInfo.GetName(); pkgName == repoOwner+"/"+repoName || strings.HasPrefix(pkgName, repoOwner+"/"+repoName+"/") {
		outputPkg.Name += "@" + tagName
		outputPkg.Version = ""
	} else {
		outputPkg.Version = tagName
	}
}

func (ctrl *Controller) getOutputtedPkg(ctx context.Context, pkg *FindingPackage, logE *logrus.Entry) *config.Package {
	outputPkg := &config.Package{
		Name:     pkg.PackageInfo.GetName(),
		Registry: pkg.RegistryName,
		Version:  "[SET PACKAGE VERSION]",
	}
	if outputPkg.Registry == "standard" {
		outputPkg.Registry = ""
	}
	if ctrl.gitHubRepositoryService == nil {
		return outputPkg
	}
	if pkgInfo := pkg.PackageInfo; pkgInfo.HasRepo() {
		ctrl.getOutputtedGitHubPkg(ctx, outputPkg, pkgInfo, logE)
		return outputPkg
	}
	return outputPkg
}
