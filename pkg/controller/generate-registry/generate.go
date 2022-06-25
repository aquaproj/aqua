package genrgst

import (
	"context"
	"fmt"
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
	github github.RepositoryService
}

func NewController(fs afero.Fs, gh github.RepositoryService) *Controller {
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
		encoder := yaml.NewEncoder(ctrl.stdout)
		encoder.SetIndent(2) //nolint:gomnd
		if err := encoder.Encode(cfg); err != nil {
			return fmt.Errorf("encode YAML: %w", err)
		}
		return nil
	}
	if err := ctrl.insert(param.InsertFile, registry.PackageInfos{pkgInfo}); err != nil {
		return err
	}
	return nil
}

func (ctrl *Controller) excludeAsset(assetName string) bool {
	suffixes := []string{
		"txt", "msi", "deb", "rpm", "md", "sig", "pem", "sbom", "apk", "dmg", "sha256",
	}
	asset := strings.ToLower(assetName)
	for _, s := range suffixes {
		if strings.HasSuffix(asset, "."+s) {
			return true
		}
	}
	words := []string{
		"readme", "license", "openbsd", "freebsd", "netbsd", "android", "386", "i386", "armv6", "armv7", "32bit",
	}
	for _, s := range words {
		if strings.Contains(asset, s) {
			return true
		}
	}
	return false
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
			assets := ctrl.listReleaseAssets(ctx, logE, pkgInfo, release.GetID())
			if len(assets) != 0 {
				assetInfos := make([]*AssetInfo, 0, len(assets))
				for _, asset := range assets {
					assetName := asset.GetName()
					if ctrl.excludeAsset(assetName) {
						continue
					}
					assetInfo := ctrl.parseAssetName(asset.GetName(), release.GetTagName())
					assetInfos = append(assetInfos, assetInfo)
				}
				ctrl.parseAssetInfos(pkgInfo, assetInfos)
			}
		}
	}
	if len(splitPkgNames) != 2 { //nolint:gomnd
		pkgInfo.Name = pkgName
	}
	return pkgInfo
}

type OS struct {
	Name string
	OS   string
}

type Arch struct {
	Name string
	Arch string
}

type AssetInfo struct {
	Template     string
	OS           string
	Arch         string
	DarwinAll    bool
	Format       string
	Replacements map[string]string
}

func has(m map[string]struct{}, key string) bool {
	_, ok := m[key]
	return ok
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

func (ctrl *Controller) getOSFormat(assetInfos []*AssetInfo, goos string) string {
	format := ""
	for _, assetInfo := range assetInfos {
		if assetInfo.OS != goos {
			continue
		}
		if assetInfo.Format != "" && assetInfo.Format != formatRaw {
			return assetInfo.Format
		}
		format = formatRaw
	}
	return format
}

func (ctrl *Controller) parseAssetInfos(pkgInfo *registry.PackageInfo, assetInfos []*AssetInfo) { //nolint:funlen,gocognit,cyclop,gocyclo
	envs := map[string]struct{}{}
	formats := map[string]int{}

	osFormats := map[string]string{}
	for _, goos := range []string{"windows", "darwin", "linux"} {
		if f := ctrl.getOSFormat(assetInfos, goos); f != "" {
			osFormats[goos] = f
		}
	}
	for _, f := range osFormats {
		formats[f]++
	}
	maxFormatCnt := 0
	if len(formats) > 1 {
		for format, cnt := range formats {
			if cnt > maxFormatCnt {
				pkgInfo.Format = format
				maxFormatCnt = cnt
				continue
			}
			if cnt == maxFormatCnt && pkgInfo.Format == formatRaw {
				pkgInfo.Format = format
				maxFormatCnt = cnt
			}
		}
	}
	for _, assetInfo := range assetInfos {
		if len(assetInfo.Replacements) != 0 {
			if pkgInfo.Replacements == nil {
				pkgInfo.Replacements = map[string]string{}
			}
			for k, v := range assetInfo.Replacements {
				if v == "pc-windows-gnu" && pkgInfo.Replacements["windows"] == "pc-windows-msvc" {
					continue
				}
				if v == "unknown-linux-gnu" && pkgInfo.Replacements["linux"] == "unknown-linux-musl" {
					continue
				}
				pkgInfo.Replacements[k] = v
			}
		}
		if assetInfo.DarwinAll {
			envs["darwin"] = struct{}{}
			continue
		}
		if assetInfo.OS == "" || assetInfo.Arch == "" {
			continue
		}
		envs[assetInfo.OS+"/"+assetInfo.Arch] = struct{}{}
		if pkgInfo.Asset == nil && assetInfo.OS != "windows" {
			if assetInfo.Format == "" || assetInfo.Format == formatRaw || len(formats) < 2 {
				pkgInfo.Asset = strP(assetInfo.Template)
			} else {
				pkgInfo.Asset = strP(strings.Replace(assetInfo.Template, "."+assetInfo.Format, ".{{.Format}}", 1))
			}
		}
		if assetInfo.OS == "linux" && assetInfo.Arch == "amd64" {
			if assetInfo.Format == "" || assetInfo.Format == formatRaw || len(formats) < 2 {
				pkgInfo.Asset = strP(assetInfo.Template)
			} else {
				pkgInfo.Asset = strP(strings.Replace(assetInfo.Template, "."+assetInfo.Format, ".{{.Format}}", 1))
			}
		}
		if pkgInfo.Format != "" { //nolint:nestif
			if pkgInfo.Format != assetInfo.Format {
				included := false
				for _, override := range pkgInfo.Overrides {
					if override.GOOS == assetInfo.OS {
						included = true
						break
					}
				}
				if !included {
					if assetInfo.Format == "raw" || pkgInfo.Format == "raw" {
						pkgInfo.Overrides = append(pkgInfo.Overrides, &registry.Override{
							GOOS:   assetInfo.OS,
							Format: assetInfo.Format,
							Asset:  strP(assetInfo.Template),
						})
					} else {
						pkgInfo.Overrides = append(pkgInfo.Overrides, &registry.Override{
							GOOS:   assetInfo.OS,
							Format: assetInfo.Format,
						})
					}
				}
			}
		}
	}
	ctrl.setSupportedEnvs(envs, pkgInfo)
}

const formatRaw = "raw"

func (ctrl *Controller) parseAssetName(assetName, version string) *AssetInfo {
	assetInfo := &AssetInfo{
		Template: strings.Replace(assetName, version, "{{.Version}}", 1),
	}
	if assetInfo.Template == assetName {
		assetInfo.Template = strings.Replace(assetName, strings.TrimPrefix(version, "v"), "{{trimV .Version}}", 1)
	}
	lowAssetName := strings.ToLower(assetName)
	ctrl.setOS(assetName, lowAssetName, assetInfo)
	ctrl.setArch(assetName, lowAssetName, assetInfo)
	if assetInfo.Arch == "" && assetInfo.OS == "darwin" {
		if strings.Contains(lowAssetName, "_all") || strings.Contains(lowAssetName, "-all") || strings.Contains(lowAssetName, ".all") {
			assetInfo.DarwinAll = true
		}
	}
	assetInfo.Format = ctrl.getFormat(assetName)
	return assetInfo
}
