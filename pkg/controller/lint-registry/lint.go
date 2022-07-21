package genrgst

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

type Controller struct {
	stdout io.Writer
	fs     afero.Fs
	github RepositoriesService
}

type RepositoriesService interface {
	Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error)
	GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
	ListReleaseAssets(ctx context.Context, owner, repo string, id int64, opts *github.ListOptions) ([]*github.ReleaseAsset, *github.Response, error)
}

func NewController(fs afero.Fs, gh RepositoriesService) *Controller {
	return &Controller{
		stdout: os.Stdout,
		fs:     fs,
		github: gh,
	}
}

func (ctrl *Controller) Run(ctx context.Context, param *config.Param, logE *logrus.Entry, args ...string) error {
	if len(args) == 0 {
		return nil
	}
	for _, arg := range args {
		if err := ctrl.lintRegistry(ctx, param, logE, arg); err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *Controller) lintRegistry(ctx context.Context, param *config.Param, logE *logrus.Entry, registryFilePath string) error {
	registryFile, err := os.Open(registryFilePath)
	if err != nil {
		return err
	}
	defer registryFile.Close()
	cfg := &registry.Config{}
	if err := yaml.NewDecoder(registryFile).Decode(cfg); err != nil {
		return err
	}
	for i, pkgInfo := range cfg.PackageInfos {
		name := pkgInfo.GetName()
		if name == "" {
			return errors.New("name is empty")
		}
		if pkgInfo.RepoOwner+"/"+pkgInfo.RepoName == pkgInfo.Name {
		}
	}
	return nil
}

type Code struct {
	ID               string
	Attributes       []string
	Level            string
	ShortDescription string
	Description      string
	URL              string
}

type Error struct {
	Indexes    []int
	Names      []string
	Code       string
	Level      string
	Attributes []string
	Value      interface{}
	Error      error
	FilePath   string
	Global     bool
}

func ListCodes() []*Code {
	return []*Code{
		{
			ID: "name_empty",
			Attributes: []string{
				"name",
				"repo_owner",
				"repo_name",
			},
			Level:            "error",
			ShortDescription: "name is empty. Either name or a pair of repo_owner and repo_name are required",
		},
		{
			ID: "name_omit",
			Attributes: []string{
				"name",
				"repo_owner",
				"repo_name",
			},
			Level:            "error",
			ShortDescription: "omit name, because name is equivalent to <repo_owner>/<repo_name>",
		},
		{
			ID: "avoid_go",
			Attributes: []string{
				"type",
			},
			Level:            "warning",
			ShortDescription: "use the package type go_install instead of go as much as possible",
		},
		{
			ID: "unknown_type",
			Attributes: []string{
				"type",
			},
			Level:            "error",
			ShortDescription: "the package type is unknown",
		},
		{
			ID: "repo_owner_miss",
			Attributes: []string{
				"repo_owner",
			},
			Level:            "error",
			ShortDescription: "repo_name is set but repo_owner isn't set",
		},
		{
			ID: "repo_name_miss",
			Attributes: []string{
				"repo_owner",
			},
			Level:            "error",
			ShortDescription: "repo_owner is set but repo_name isn't set",
		},
		{
			ID: "asset_miss",
			Attributes: []string{
				"asset",
			},
			Level:            "error",
			ShortDescription: "asset is required for github_release package",
		},
		{
			ID: "asset_unneeded",
			Attributes: []string{
				"asset",
			},
			Level:            "error",
			ShortDescription: "asset is unneeded",
		},
		{
			ID: "path_miss",
			Attributes: []string{
				"path",
			},
			Level:            "error",
			ShortDescription: "path is required for github_content or go_install package",
		},
		{
			ID: "path_unneeded",
			Attributes: []string{
				"path",
			},
			Level:            "error",
			ShortDescription: "path is unneeded",
		},
		{
			ID: "format_unsupported",
			Attributes: []string{
				"format",
			},
			Level:            "error",
			ShortDescription: "format value is unsupported",
		},
		{
			ID: "description_empty",
			Attributes: []string{
				"description",
			},
			Level:            "error",
			ShortDescription: "description is empty",
		},
		{
			ID: "description_trim_space",
			Attributes: []string{
				"description",
			},
			Level:            "error",
			ShortDescription: "remove all leading and trailing white space from description",
		},
		{
			ID: "description_punctuation",
			Attributes: []string{
				"description",
			},
			Level:            "error",
			ShortDescription: "remove punctuation from the end of description",
		},
		{
			ID: "link_empty",
			Attributes: []string{
				"link",
			},
			Level:            "error",
			ShortDescription: "link is empty",
		},
		{
			ID: "link_empty",
			Attributes: []string{
				"link",
			},
			Level:            "error",
			ShortDescription: "link is empty",
		},
	}
}

type LintPackageFunc func(param *config.Param, pkgInfo *registry.PackageInfo) *Error

func (ctrl *Controller) lintPackage(param *config.Param, pkgInfo *registry.PackageInfo) []*Error {
	funcs := []LintPackageFunc{
		lintName,
		lintType,
		lintRepo,
		lintAsset,
		lintPath,
		lintFormat,
		lintFiles,
		lintURL,
		lintDescription,
		lintLink,
		lintReplacements,
		lintOverrides,
		lintFormatOverrides,
		lintVersionConstraint,
		lintVersionOverrides,
		lintSupportedIf,
		lintSupportedEnvs,
		lintVersionFilter,
		lintRosetta2,
		lintAliases,
		lintVersionSource,
		lintCompleteWindowsExt,
		lintWindowsExt,
		lintSearchWords,
	}
	errs := []*Error{}
	for _, fn := range funcs {
		if e := fn(param, pkgInfo); e != nil {
			errs = append(errs, e)
		}
	}
	return errs
}

func lintName(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	name := pkgInfo.GetName()
	if name == "" {
		return &Error{
			Code: "name_empty",
		}
	}
	if pkgInfo.RepoOwner+"/"+pkgInfo.RepoName == pkgInfo.Name {
		return &Error{
			Code: "name_omit",
		}
	}
	return nil
}

func lintType(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	switch pkgInfo.Type {
	case registry.PkgInfoTypeGitHubRelease, registry.PkgInfoTypeGitHubContent, registry.PkgInfoTypeGitHubArchive, registry.PkgInfoTypeHTTP, registry.PkgInfoTypeGoInstall:
		return nil
	case registry.PkgInfoTypeGo:
		return &Error{
			Code: "avoid_go",
		}
	}
	return &Error{
		Code: "unknown_type",
	}
}

func lintRepo(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	if pkgInfo.RepoOwner == "" {
		if pkgInfo.RepoName == "" {
			return nil
		}
		return &Error{
			Code: "repo_owner_miss",
		}
	}
	if pkgInfo.RepoName == "" {
		return &Error{
			Code: "repo_name_miss",
		}
	}
	return nil
}

func lintAsset(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	if pkgInfo.Type == registry.PkgInfoTypeGitHubRelease {
		if pkgInfo.Asset == nil {
			return &Error{
				Code: "asset_miss",
			}
		}
		if *pkgInfo.Asset == "" {
			return &Error{
				Code: "asset_miss",
			}
		}
		return nil
	}
	if pkgInfo.Asset != nil {
		return &Error{
			Code: "asset_unneeded",
		}
	}
	return nil
}

func lintPath(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	p := pkgInfo.GetPath()
	switch pkgInfo.Type {
	case registry.PkgInfoTypeGitHubContent, registry.PkgInfoTypeGoInstall:
		if p == "" {
			return &Error{
				Code: "path_miss",
			}
		}
		return nil
	}
	if p != "" {
		return &Error{
			Code: "path_unneeded",
		}
	}
	return nil
}

func lintFormat(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	return nil
}

func lintFiles(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	return nil
}

func lintURL(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	return nil
}

func lintDescription(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	if pkgInfo.Description == "" {
		return &Error{
			Code: "description_empty",
		}
	}
	if strings.TrimSpace(pkgInfo.Description) != pkgInfo.Description {
		return &Error{
			Code: "description_trim_space",
		}
	}
	if strings.TrimRight(strings.TrimSpace(pkgInfo.Description), ",.!?") != pkgInfo.Description {
		return &Error{
			Code: "description_punctuation",
		}
	}
	return nil
}

func lintLink(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	if pkgInfo.Link == "" {
		if pkgInfo.HasRepo() {
			return nil
		}
		return &Error{
			Code: "link_empty",
		}
	}
	if pkgInfo.HasRepo() {
		return &Error{
			Code: "link_unneeded",
		}
	}
	return nil
}

func lintReplacements(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	return nil
}

func lintOverrides(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	return nil
}

func lintFormatOverrides(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	if pkgInfo.FormatOverrides != nil {
		if pkgInfo.Overrides == nil {
			return &Error{
				Code: "format_overrides_deprecated",
			}
		}
		return &Error{
			Code:  "format_overrides_deprecated",
			Level: "warning",
		}
	}
	return nil
}

func lintVersionConstraint(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	return nil
}

func lintVersionOverrides(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	return nil
}

func lintSupportedIf(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	if pkgInfo.SupportedIf != nil {
		return &Error{
			Code: "supported_if_deprecated",
		}
	}
	return nil
}

func lintSupportedEnvs(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	return nil
}

func lintVersionFilter(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	return nil
}

func lintRosetta2(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	return nil
}

func lintAliases(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	if pkgInfo.Aliases != nil {
		return &Error{
			Code: "aliases_use_only_compatibility",
		}
	}
	return nil
}

func lintVersionSource(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	switch pkgInfo.VersionSource {
	case "", "github_tag":
		return nil
	}
	return &Error{
		Code: "github_tag_unknown",
	}
}

func lintCompleteWindowsExt(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	return nil
}

func lintWindowsExt(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	return nil
}

func lintSearchWords(param *config.Param, pkgInfo *registry.PackageInfo) *Error {
	return nil
}
