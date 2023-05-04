package genrgst

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/asset"
	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/generate/output"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/forPelevin/gomoji"
	yaml "github.com/goccy/go-yaml"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Controller struct {
	stdout            io.Writer
	fs                afero.Fs
	github            RepositoriesService
	testdataOutputter TestdataOutputter
}

type TestdataOutputter interface {
	Output(param *output.Param) error
}

func NewController(fs afero.Fs, gh RepositoriesService, testdataOutputter TestdataOutputter) *Controller {
	return &Controller{
		stdout:            os.Stdout,
		fs:                fs,
		github:            gh,
		testdataOutputter: testdataOutputter,
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
	pkgInfo, versions := ctrl.getPackageInfo(ctx, logE, pkgName, param.Deep)
	if param.OutTestData != "" {
		if err := ctrl.testdataOutputter.Output(&output.Param{
			List: listPkgsFromVersions(pkgName, versions),
			Dest: param.OutTestData,
		}); err != nil {
			return fmt.Errorf("output testdata to a file: %w", err)
		}
	}
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

func (ctrl *Controller) getRelease(ctx context.Context, repoOwner, repoName, version string) (*github.RepositoryRelease, error) {
	if version == "" {
		release, _, err := ctrl.github.GetLatestRelease(ctx, repoOwner, repoName)
		return release, err //nolint:wrapcheck
	}
	release, _, err := ctrl.github.GetReleaseByTag(ctx, repoOwner, repoName, version)
	return release, err //nolint:wrapcheck
}

func (ctrl *Controller) getPackageInfo(ctx context.Context, logE *logrus.Entry, arg string, deep bool) (*registry.PackageInfo, []string) {
	pkgName, version, _ := strings.Cut(arg, "@")
	splitPkgNames := strings.Split(pkgName, "/")
	pkgInfo := &registry.PackageInfo{
		Type: "github_release",
	}
	if len(splitPkgNames) == 1 {
		pkgInfo.Name = pkgName
		return pkgInfo, nil
	}
	if len(splitPkgNames) != 2 { //nolint:gomnd
		pkgInfo.Name = pkgName
	}
	pkgInfo.RepoOwner = splitPkgNames[0]
	pkgInfo.RepoName = splitPkgNames[1]
	repo, _, err := ctrl.github.Get(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName)
	if err != nil {
		logE.WithFields(logrus.Fields{
			"repo_owner": pkgInfo.RepoOwner,
			"repo_name":  pkgInfo.RepoName,
		}).WithError(err).Warn("get the repository")
	} else {
		pkgInfo.Description = strings.TrimRight(strings.TrimSpace(gomoji.RemoveEmojis(repo.GetDescription())), ".!?")
	}
	if deep && version == "" {
		return ctrl.getPackageInfoWithVersionOverrides(ctx, logE, pkgName, pkgInfo)
	}
	release, err := ctrl.getRelease(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, version)
	if err != nil {
		logE.WithFields(logrus.Fields{
			"repo_owner": pkgInfo.RepoOwner,
			"repo_name":  pkgInfo.RepoName,
		}).WithError(err).Warn("get the release")
	} else {
		logE.WithField("version", release.GetTagName()).Debug("got the release")
		assets := ctrl.listReleaseAssets(ctx, logE, pkgInfo, release.GetID())
		logE.WithField("num_of_assets", len(assets)).Debug("got assets")
		ctrl.patchRelease(logE, pkgInfo, pkgName, release.GetTagName(), assets)
	}
	return pkgInfo, []string{version}
}

func (ctrl *Controller) patchRelease(logE *logrus.Entry, pkgInfo *registry.PackageInfo, pkgName string, tagName string, assets []*github.ReleaseAsset) { //nolint:funlen,cyclop
	if len(assets) == 0 {
		return
	}
	assetInfos := make([]*asset.AssetInfo, 0, len(assets))
	pkgNameContainChecksum := strings.Contains(strings.ToLower(pkgName), "checksum")
	assetNames := map[string]struct{}{}
	checksumNames := map[string]struct{}{}
	for _, aset := range assets {
		assetName := aset.GetName()
		if !pkgNameContainChecksum {
			chksum := checksum.GetChecksumConfigFromFilename(assetName, tagName)
			if chksum != nil {
				checksumNames[assetName] = struct{}{}
				continue
			}
		}
		if asset.Exclude(pkgName, assetName, tagName) {
			logE.WithField("asset_name", assetName).Debug("exclude an asset")
			continue
		}
		assetNames[assetName] = struct{}{}
		assetInfo := asset.ParseAssetName(assetName, tagName)
		assetInfos = append(assetInfos, assetInfo)
	}
	for assetName := range assetNames {
		if _, ok := checksumNames[assetName+".md5"]; ok {
			pkgInfo.Checksum = &registry.Checksum{
				Type:       "github_release",
				Asset:      "{{.Asset}}.md5",
				FileFormat: "regexp",
				Algorithm:  "md5",
				Pattern: &registry.ChecksumPattern{
					Checksum: `^(\b[A-Fa-f0-9]{32}\b)`,
					File:     `^\b[A-Fa-f0-9]{32}\b\s+(\S+)$`,
				},
			}
			break
		}
		if _, ok := checksumNames[assetName+".sha256"]; ok {
			pkgInfo.Checksum = &registry.Checksum{
				Type:       "github_release",
				Asset:      "{{.Asset}}.sha256",
				FileFormat: "regexp",
				Algorithm:  "sha256",
				Pattern: &registry.ChecksumPattern{
					Checksum: `^(\b[A-Fa-f0-9]{64}\b)`,
					File:     `^\b[A-Fa-f0-9]{64}\b\s+(\S+)$`,
				},
			}
			break
		}
		if _, ok := checksumNames[assetName+".sha512"]; ok {
			pkgInfo.Checksum = &registry.Checksum{
				Type:       "github_release",
				Asset:      "{{.Asset}}.sha512",
				FileFormat: "regexp",
				Algorithm:  "sha512",
				Pattern: &registry.ChecksumPattern{
					Checksum: `^(\b[A-Fa-f0-9]{128}\b)`,
					File:     `^\b[A-Fa-f0-9]{128}\b\s+(\S+)$`,
				},
			}
			break
		}
		if _, ok := checksumNames[assetName+".sha1"]; ok {
			pkgInfo.Checksum = &registry.Checksum{
				Type:       "github_release",
				Asset:      "{{.Asset}}.sha512",
				FileFormat: "regexp",
				Algorithm:  "sha1",
				Pattern: &registry.ChecksumPattern{
					Checksum: `^(\b[A-Fa-f0-9]{40}\b)`,
					File:     `^\b[A-Fa-f0-9]{40}\b\s+(\S+)$`,
				},
			}
			break
		}
	}
	if len(checksumNames) > 0 && pkgInfo.Checksum == nil {
		for checksumName := range checksumNames {
			chksum := checksum.GetChecksumConfigFromFilename(checksumName, tagName)
			if chksum != nil {
				assetInfo := asset.ParseAssetName(checksumName, tagName)
				chksum.Asset = assetInfo.Template
				pkgInfo.Checksum = chksum
				break
			}
		}
	}
	asset.ParseAssetInfos(pkgInfo, assetInfos)
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
