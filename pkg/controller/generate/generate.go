package generate

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/antonmedv/expr/vm"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/expr"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"gopkg.in/yaml.v2"
)

type Controller struct {
	stdin             io.Reader
	stdout            io.Writer
	github            RepositoriesService
	registryInstaller domain.RegistryInstaller
	configFinder      ConfigFinder
	configReader      domain.ConfigReader
	fuzzyFinder       FuzzyFinder
	versionSelector   VersionSelector
	fs                afero.Fs
}

type RepositoriesService interface {
	GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
	ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error)
	ListTags(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error)
}

type ConfigFinder interface {
	Find(wd, configFilePath string, globalConfigFilePaths ...string) (string, error)
}

func New(configFinder ConfigFinder, configReader domain.ConfigReader, registInstaller domain.RegistryInstaller, gh RepositoriesService, fs afero.Fs, fuzzyFinder FuzzyFinder, versionSelector VersionSelector) *Controller {
	return &Controller{
		stdin:             os.Stdin,
		stdout:            os.Stdout,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registInstaller,
		github:            gh,
		fs:                fs,
		fuzzyFinder:       fuzzyFinder,
		versionSelector:   versionSelector,
	}
}

// Generate searches packages in registries and outputs the configuration to standard output.
// If no package is specified, the interactive fuzzy finder is launched.
// If the package supports, the latest version is gotten by GitHub API.
func (ctrl *Controller) Generate(ctx context.Context, logE *logrus.Entry, param *config.Param, args ...string) error {
	cfgFilePath, err := ctrl.configFinder.Find(param.PWD, param.ConfigFilePath, param.GlobalConfigFilePaths...)
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
	if !param.Insert && param.Dest == "" {
		if err := yaml.NewEncoder(ctrl.stdout).Encode(list); err != nil {
			return fmt.Errorf("output generated package configuration: %w", err)
		}
		return nil
	}

	if param.Dest != "" {
		if _, err := ctrl.fs.Stat(param.Dest); err != nil {
			if err := afero.WriteFile(ctrl.fs, param.Dest, []byte("packages:\n\n"), 0o644); err != nil { //nolint:gomnd
				return fmt.Errorf("create a file: %w", err)
			}
		}
		return ctrl.generateInsert(param.Dest, list)
	}

	return ctrl.generateInsert(cfgFilePath, list)
}

type FindingPackage struct {
	PackageInfo  *registry.PackageInfo
	RegistryName string
}

func (ctrl *Controller) generate(ctx context.Context, logE *logrus.Entry, param *config.Param, cfgFilePath string, args ...string) ([]*aqua.Package, error) {
	cfg := &aqua.Config{}
	if err := ctrl.configReader.Read(cfgFilePath, cfg); err != nil {
		return nil, err //nolint:wrapcheck
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
	idxes, err := ctrl.fuzzyFinder.Find(pkgs)
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil, nil
		}
		return nil, fmt.Errorf("find the package: %w", err)
	}
	arr := make([]*aqua.Package, len(idxes))
	for i, idx := range idxes {
		arr[i] = ctrl.getOutputtedPkg(ctx, param, pkgs[idx], logE)
	}

	return arr, nil
}

func getGeneratePkg(s string) string {
	if !strings.Contains(s, ",") {
		return "standard," + s
	}
	return s
}

func (ctrl *Controller) outputListedPkgs(ctx context.Context, logE *logrus.Entry, param *config.Param, registryContents map[string]*registry.Config, pkgNames ...string) ([]*aqua.Package, error) {
	m := map[string]*FindingPackage{}
	for registryName, registryContent := range registryContents {
		logE := logE.WithField("registry_name", registryName)
		for pkgName, pkg := range registryContent.PackageInfos.ToMapWarn(logE) {
			pkg := pkg
			logE := logE.WithField("package_name", pkgName)
			m[registryName+","+pkgName] = &FindingPackage{
				PackageInfo:  pkg,
				RegistryName: registryName,
			}
			for _, alias := range pkg.Aliases {
				if alias.Name == "" {
					logE.Warn("ignore a package alias because the alias is empty")
					continue
				}
				m[registryName+","+alias.Name] = &FindingPackage{
					PackageInfo:  pkg,
					RegistryName: registryName,
				}
			}
		}
	}

	outputPkgs := []*aqua.Package{}
	for _, pkgName := range pkgNames {
		pkgName = getGeneratePkg(pkgName)
		findingPkg, ok := m[pkgName]
		if !ok {
			return nil, logerr.WithFields(errUnknownPkg, logrus.Fields{"package_name": pkgName}) //nolint:wrapcheck
		}
		outputPkg := ctrl.getOutputtedPkg(ctx, param, findingPkg, logE)
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

func (ctrl *Controller) readGeneratedPkgsFromFile(ctx context.Context, param *config.Param, outputPkgs []*aqua.Package, m map[string]*FindingPackage, logE *logrus.Entry) ([]*aqua.Package, error) {
	var file io.Reader
	if param.File == "-" {
		file = ctrl.stdin
	} else {
		f, err := ctrl.fs.Open(param.File)
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
		outputPkg := ctrl.getOutputtedPkg(ctx, param, findingPkg, logE)
		outputPkgs = append(outputPkgs, outputPkg)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read the file: %w", err)
	}
	return outputPkgs, nil
}

func (ctrl *Controller) listTags(ctx context.Context, pkgInfo *registry.PackageInfo, logE *logrus.Entry) []*github.RepositoryTag {
	repoOwner := pkgInfo.RepoOwner
	repoName := pkgInfo.RepoName
	opt := &github.ListOptions{
		PerPage: 100, //nolint:gomnd
	}
	var versionFilter *vm.Program
	if pkgInfo.VersionFilter != nil {
		var err error
		versionFilter, err = expr.CompileVersionFilter(*pkgInfo.VersionFilter)
		if err != nil {
			return nil
		}
	}
	var arr []*github.RepositoryTag
	for i := 0; i < 10; i++ {
		tags, _, err := ctrl.github.ListTags(ctx, repoOwner, repoName, opt)
		if err != nil {
			logerr.WithError(logE, err).WithFields(logrus.Fields{
				"repo_owner": repoOwner,
				"repo_name":  repoName,
			}).Warn("list releases")
			return arr
		}
		for _, tag := range tags {
			if versionFilter != nil {
				f, err := expr.EvaluateVersionFilter(versionFilter, tag.GetName())
				if err != nil || !f {
					continue
				}
			}
			arr = append(arr, tag)
		}
		if len(tags) != opt.PerPage {
			return arr
		}
		opt.Page++
	}
	return arr
}

func (ctrl *Controller) listReleases(ctx context.Context, pkgInfo *registry.PackageInfo, logE *logrus.Entry) []*github.RepositoryRelease { //nolint:cyclop
	repoOwner := pkgInfo.RepoOwner
	repoName := pkgInfo.RepoName
	opt := &github.ListOptions{
		PerPage: 100, //nolint:gomnd
	}
	var versionFilter *vm.Program
	if pkgInfo.VersionFilter != nil {
		var err error
		versionFilter, err = expr.CompileVersionFilter(*pkgInfo.VersionFilter)
		if err != nil {
			return nil
		}
	}
	var arr []*github.RepositoryRelease
	for i := 0; i < 10; i++ {
		releases, _, err := ctrl.github.ListReleases(ctx, repoOwner, repoName, opt)
		if err != nil {
			logerr.WithError(logE, err).WithFields(logrus.Fields{
				"repo_owner": repoOwner,
				"repo_name":  repoName,
			}).Warn("list releases")
			return arr
		}
		for _, release := range releases {
			if release.GetPrerelease() {
				continue
			}
			if versionFilter != nil {
				f, err := expr.EvaluateVersionFilter(versionFilter, release.GetTagName())
				if err != nil || !f {
					continue
				}
			}
			arr = append(arr, release)
		}
		if len(releases) != opt.PerPage {
			return arr
		}
		opt.Page++
	}
	return arr
}

func (ctrl *Controller) listAndGetTagName(ctx context.Context, pkgInfo *registry.PackageInfo, logE *logrus.Entry) string {
	repoOwner := pkgInfo.RepoOwner
	repoName := pkgInfo.RepoName
	opt := &github.ListOptions{
		PerPage: 30, //nolint:gomnd
	}
	versionFilter, err := expr.CompileVersionFilter(*pkgInfo.VersionFilter)
	if err != nil {
		return ""
	}
	for {
		releases, _, err := ctrl.github.ListReleases(ctx, repoOwner, repoName, opt)
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
			f, err := expr.EvaluateVersionFilter(versionFilter, release.GetTagName())
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

func (ctrl *Controller) listAndGetTagNameFromTag(ctx context.Context, pkgInfo *registry.PackageInfo, logE *logrus.Entry) string {
	repoOwner := pkgInfo.RepoOwner
	repoName := pkgInfo.RepoName
	opt := &github.ListOptions{
		PerPage: 30, //nolint:gomnd
	}
	versionFilter, err := expr.CompileVersionFilter(*pkgInfo.VersionFilter)
	if err != nil {
		return ""
	}
	for {
		tags, _, err := ctrl.github.ListTags(ctx, repoOwner, repoName, opt)
		if err != nil {
			logerr.WithError(logE, err).WithFields(logrus.Fields{
				"repo_owner": repoOwner,
				"repo_name":  repoName,
			}).Warn("list releases")
			return ""
		}
		for _, tag := range tags {
			tagName := tag.GetName()
			f, err := expr.EvaluateVersionFilter(versionFilter, tagName)
			if err != nil || !f {
				continue
			}
			return tagName
		}
		if len(tags) != opt.PerPage {
			return ""
		}
		opt.Page++
	}
}

func (ctrl *Controller) getOutputtedGitHubPkgFromTag(ctx context.Context, param *config.Param, outputPkg *aqua.Package, pkgInfo *registry.PackageInfo, logE *logrus.Entry) {
	repoOwner := pkgInfo.RepoOwner
	repoName := pkgInfo.RepoName
	var tagName string

	if param.SelectVersion { //nolint:nestif
		tags := ctrl.listTags(ctx, pkgInfo, logE)
		versions := make([]*Version, len(tags))
		for i, tag := range tags {
			versions[i] = &Version{
				Name:    tag.GetName(),
				Version: tag.GetName(),
			}
		}
		idx, err := ctrl.versionSelector.Find(versions)
		if err != nil {
			return
		}
		tagName = versions[idx].Version
	} else {
		if pkgInfo.VersionFilter != nil {
			tagName = ctrl.listAndGetTagNameFromTag(ctx, pkgInfo, logE)
		} else {
			tags, _, err := ctrl.github.ListTags(ctx, repoOwner, repoName, nil)
			if err != nil {
				logerr.WithError(logE, err).WithFields(logrus.Fields{
					"repo_owner": repoOwner,
					"repo_name":  repoName,
				}).Warn("list GitHub tags")
				return
			}
			if len(tags) == 0 {
				return
			}
			tag := tags[0]
			tagName = tag.GetName()
		}
	}

	if pkgName := pkgInfo.GetName(); pkgName == repoOwner+"/"+repoName || strings.HasPrefix(pkgName, repoOwner+"/"+repoName+"/") {
		outputPkg.Name += "@" + tagName
		outputPkg.Version = ""
	} else {
		outputPkg.Version = tagName
	}
}

func (ctrl *Controller) getOutputtedGitHubPkg(ctx context.Context, param *config.Param, outputPkg *aqua.Package, pkgInfo *registry.PackageInfo, logE *logrus.Entry) {
	if pkgInfo.VersionSource == "github_tag" {
		ctrl.getOutputtedGitHubPkgFromTag(ctx, param, outputPkg, pkgInfo, logE)
		return
	}
	repoOwner := pkgInfo.RepoOwner
	repoName := pkgInfo.RepoName
	var tagName string

	if param.SelectVersion { //nolint:nestif
		releases := ctrl.listReleases(ctx, pkgInfo, logE)
		versions := make([]*Version, len(releases))
		for i, release := range releases {
			versions[i] = &Version{
				Name:        release.GetName(),
				Version:     release.GetTagName(),
				Description: release.GetBody(),
				URL:         release.GetHTMLURL(),
			}
		}
		idx, err := ctrl.versionSelector.Find(versions)
		if err != nil {
			return
		}
		tagName = versions[idx].Version
	} else {
		if pkgInfo.VersionFilter != nil {
			tagName = ctrl.listAndGetTagName(ctx, pkgInfo, logE)
		} else {
			release, _, err := ctrl.github.GetLatestRelease(ctx, repoOwner, repoName)
			if err != nil {
				logerr.WithError(logE, err).WithFields(logrus.Fields{
					"repo_owner": repoOwner,
					"repo_name":  repoName,
				}).Warn("get the latest release")
				return
			}
			tagName = release.GetTagName()
		}
	}

	if pkgName := pkgInfo.GetName(); pkgName == repoOwner+"/"+repoName || strings.HasPrefix(pkgName, repoOwner+"/"+repoName+"/") {
		outputPkg.Name += "@" + tagName
		outputPkg.Version = ""
	} else {
		outputPkg.Version = tagName
	}
}

func (ctrl *Controller) getOutputtedPkg(ctx context.Context, param *config.Param, pkg *FindingPackage, logE *logrus.Entry) *aqua.Package {
	outputPkg := &aqua.Package{
		Name:     pkg.PackageInfo.GetName(),
		Registry: pkg.RegistryName,
		Version:  "[SET PACKAGE VERSION]",
	}
	if outputPkg.Registry == "standard" {
		outputPkg.Registry = ""
	}
	if ctrl.github == nil {
		return outputPkg
	}
	if pkgInfo := pkg.PackageInfo; pkgInfo.HasRepo() {
		ctrl.getOutputtedGitHubPkg(ctx, param, outputPkg, pkgInfo, logE)
		return outputPkg
	}
	return outputPkg
}
