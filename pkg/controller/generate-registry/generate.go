package genrgst

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aquaproj/aqua/pkg/asset"
	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/goccy/go-yaml"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
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

func (ctrl *Controller) GenerateRegistry(ctx context.Context, param *config.Param, logE *logrus.Entry, args ...string) error {
	if len(args) == 0 {
		return nil
	}
	for _, arg := range args {
		if err := ctrl.genRegistry(ctx, param, logE, arg); err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *Controller) genRegistry(ctx context.Context, param *config.Param, logE *logrus.Entry, pkgName string) error {
	pkgInfo := ctrl.getPackageInfo(ctx, logE, pkgName)
	if param.InsertFile == "" {
		cfg := &registry.Config{
			PackageInfos: registry.PackageInfos{
				pkgInfo,
			},
		}
		encoder := yaml.NewEncoder(ctrl.stdout, yaml.IndentSequence(true))
		if err := encoder.EncodeContext(ctx, cfg); err != nil {
			return fmt.Errorf("encode YAML: %w", err)
		}
		return nil
	}
	if err := ctrl.insert(param.InsertFile, registry.PackageInfos{pkgInfo}); err != nil {
		return err
	}
	return nil
}

func (ctrl *Controller) getPackageInfo(ctx context.Context, logE *logrus.Entry, pkgName string) *registry.PackageInfo {
	splitPkgNames := strings.Split(pkgName, "/")
	pkgInfo := &registry.PackageInfo{
		Type: "github_release",
	}
	if len(splitPkgNames) > 1 { //nolint:nestif
		pkgInfo.RepoOwner = splitPkgNames[0]
		pkgInfo.RepoName = splitPkgNames[1]
		repo, _, err := ctrl.github.Get(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName)
		if err != nil {
			logE.WithFields(logrus.Fields{
				"repo_owner": pkgInfo.RepoOwner,
				"repo_name":  pkgInfo.RepoName,
			}).WithError(err).Warn("get the repository")
		} else {
			pkgInfo.Description = strings.TrimRight(strings.TrimSpace(repo.GetDescription()), ".!?")
		}
		release, _, err := ctrl.github.GetLatestRelease(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName)
		if err != nil {
			logE.WithFields(logrus.Fields{
				"repo_owner": pkgInfo.RepoOwner,
				"repo_name":  pkgInfo.RepoName,
			}).WithError(err).Warn("get the latest release")
		} else {
			logE.WithField("version", release.GetTagName()).Debug("got the latest release")
			assets := ctrl.listReleaseAssets(ctx, logE, pkgInfo, release.GetID())
			if len(assets) != 0 {
				logE.WithField("num_of_assets", len(assets)).Debug("got assets")
				assetInfos := make([]*asset.AssetInfo, 0, len(assets))
				pkgNameContainChecksum := strings.Contains(strings.ToLower(pkgName), "checksum")
				for _, aset := range assets {
					assetName := aset.GetName()
					if !pkgNameContainChecksum {
						chksum := checksum.GetChecksumConfigFromFilename(assetName, release.GetTagName())
						if chksum != nil {
							pkgInfo.Checksum = chksum
							continue
						}
					}
					if asset.Exclude(pkgName, assetName, release.GetTagName()) {
						logE.WithField("asset_name", assetName).Debug("exclude an asset")
						continue
					}
					assetInfo := asset.ParseAssetName(aset.GetName(), release.GetTagName())
					assetInfos = append(assetInfos, assetInfo)
				}
				asset.ParseAssetInfos(pkgInfo, assetInfos)
			}
		}
	}
	if len(splitPkgNames) != 2 { //nolint:gomnd
		pkgInfo.Name = pkgName
	}
	return pkgInfo
}

func boolP(b bool) *bool {
	return &b
}

func strP(s string) *string {
	return &s
}

func (ctrl *Controller) listReleaseAssets(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo, releaseID int64) []*github.ReleaseAsset {
	opts := &github.ListOptions{
		PerPage: 100, //nolint:gomnd
	}
	var arr []*github.ReleaseAsset
	for i := 0; i < 10; i++ {
		assets, _, err := ctrl.github.ListReleaseAssets(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, releaseID, opts)
		if err != nil {
			logE.WithFields(logrus.Fields{
				"repo_owner": pkgInfo.RepoOwner,
				"repo_name":  pkgInfo.RepoName,
			}).WithError(err).Warn("list release assets")
			return arr
		}
		arr = append(arr, assets...)
		if len(assets) < opts.PerPage {
			return arr
		}
		opts.Page++
	}
	return arr
}
