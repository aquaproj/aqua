package genrgst

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/mholt/archiver/v3"
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
		"txt", "msi", "deb", "rpm", "md", "sig", "pem", "sbom", "apk",
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

func (ctrl *Controller) parseAssetInfos(pkgInfo *registry.PackageInfo, assetInfos []*AssetInfo) { //nolint:funlen,gocognit,cyclop
	envs := map[string]struct{}{}
	formats := map[string]int{}
	for _, assetInfo := range assetInfos {
		if assetInfo.Format != "" {
			formats[assetInfo.Format]++
		}
	}
	maxFormatCnt := 0
	if len(formats) > 1 {
		for format, cnt := range formats {
			if cnt > maxFormatCnt {
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
				pkgInfo.Asset = strP(strings.Replace(assetInfo.Template, "."+assetInfo.Format, ".{{Format}}", 1))
			}
		}
		if assetInfo.OS == "linux" && assetInfo.Arch == "amd64" {
			if assetInfo.Format == "" || assetInfo.Format == formatRaw || len(formats) < 2 {
				pkgInfo.Asset = strP(assetInfo.Template)
			} else {
				pkgInfo.Asset = strP(strings.Replace(assetInfo.Template, "."+assetInfo.Format, ".{{Format}}", 1))
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
					pkgInfo.Overrides = append(pkgInfo.Overrides, &registry.Override{
						GOOS:   assetInfo.OS,
						Format: assetInfo.Format,
					})
				}
			}
		}
	}
	ctrl.setSupportedEnvs(envs, pkgInfo)
}

func (ctrl *Controller) setSupportedEnvs(envs map[string]struct{}, pkgInfo *registry.PackageInfo) { //nolint:cyclop
	if has(envs, "darwin") || has(envs, "darwin/amd64") {
		pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, "darwin")
	} else if has(envs, "darwin/arm64") {
		pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, "darwin/arm64")
	}
	if has(envs, "darwin/amd64") && !has(envs, "darwin/arm64") {
		pkgInfo.Rosetta2 = boolP(true)
	}
	if has(envs, "linux/amd64") {
		if has(envs, "linux/arm64") {
			pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, "linux")
		} else {
			pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, "linux/amd64")
		}
	} else if has(envs, "linux/arm64") {
		pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, "linux/arm64")
	}
	if has(envs, "windows/amd64") {
		if has(envs, "windows/arm64") {
			pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, "windows")
		} else {
			pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, "windows/amd64")
		}
	} else if has(envs, "windows/arm64") {
		pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, "windows/arm64")
	}
	if reflect.DeepEqual(pkgInfo.SupportedEnvs, []string{"darwin", "linux", "windows"}) {
		pkgInfo.SupportedEnvs = nil
	}
	if reflect.DeepEqual(pkgInfo.SupportedEnvs, []string{"darwin", "linux", "windows/amd64"}) {
		pkgInfo.SupportedEnvs = []string{"darwin", "linux", "amd64"}
	}
	if reflect.DeepEqual(pkgInfo.SupportedEnvs, []string{"darwin", "linux/amd64", "windows/amd64"}) {
		pkgInfo.SupportedEnvs = []string{"darwin", "amd64"}
	}
}

func (ctrl *Controller) setOS(assetName, lowAssetName string, assetInfo *AssetInfo) { //nolint:funlen
	osList := []*OS{
		{
			Name: "apple-darwin",
			OS:   "darwin",
		},
		{
			Name: "unknown-linux-gnu",
			OS:   "linux",
		},
		{
			Name: "unknown-linux-musl",
			OS:   "linux",
		},
		{
			Name: "pc-windows-msvc",
			OS:   "windows",
		},
		{
			Name: "pc-windows-gnu",
			OS:   "windows",
		},
		{
			Name: "darwin",
			OS:   "darwin",
		},
		{
			Name: "linux",
			OS:   "linux",
		},
		{
			Name: "windows",
			OS:   "windows",
		},
		{
			Name: "apple",
			OS:   "darwin",
		},
		{
			Name: "osx",
			OS:   "darwin",
		},
		{
			Name: "macos",
			OS:   "darwin",
		},
		{
			Name: "mac",
			OS:   "darwin",
		},
		{
			Name: "win64",
			OS:   "windows",
		},
		{
			Name: "win",
			OS:   "windows",
		},
	}
	for _, o := range osList {
		if idx := strings.Index(lowAssetName, o.Name); idx != -1 {
			osName := assetName[idx : idx+len(o.Name)]
			assetInfo.OS = o.OS
			if osName != o.OS {
				if assetInfo.Replacements == nil {
					assetInfo.Replacements = map[string]string{}
				}
				assetInfo.Replacements[o.OS] = osName
			}
			assetInfo.Template = strings.Replace(assetInfo.Template, osName, "{{.OS}}", 1)
			break
		}
	}
	if assetInfo.OS == "" && strings.Contains(lowAssetName, ".exe") {
		assetInfo.OS = "windows"
	}
}

func (ctrl *Controller) setArch(assetName, lowAssetName string, assetInfo *AssetInfo) {
	archList := []*Arch{
		{
			Name: "amd64",
			Arch: "amd64",
		},
		{
			Name: "arm64",
			Arch: "arm64",
		},
		{
			Name: "x86_64",
			Arch: "amd64",
		},
		{
			Name: "64bit",
			Arch: "amd64",
		},
		{
			Name: "aarch64",
			Arch: "arm64",
		},
	}
	for _, o := range archList {
		if idx := strings.Index(lowAssetName, o.Name); idx != -1 {
			archName := assetName[idx : idx+len(o.Name)]
			assetInfo.Arch = o.Arch
			if archName != o.Arch {
				if assetInfo.Replacements == nil {
					assetInfo.Replacements = map[string]string{}
				}
				assetInfo.Replacements[o.Arch] = archName
			}
			assetInfo.Template = strings.Replace(assetInfo.Template, archName, "{{.Arch}}", 1)
			break
		}
	}
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

func (ctrl *Controller) getFormat(assetName string) string { //nolint:funlen,cyclop
	a, err := archiver.ByExtension(assetName)
	if err != nil {
		return formatRaw
	}
	switch a.(type) {
	case *archiver.Rar:
		return "rar"
	case *archiver.Tar:
		return "tar"
	case *archiver.TarBrotli:
		if strings.HasSuffix(assetName, ".tbr") {
			return "tbr"
		}
		return "tar.br"
	case *archiver.TarBz2:
		if strings.HasSuffix(assetName, ".tbz2") {
			return "btz2"
		}
		return "tar.bz2"
	case *archiver.TarGz:
		if strings.HasSuffix(assetName, ".tgz") {
			return "tgz"
		}
		return "tar.gz"
	case *archiver.TarLz4:
		if strings.HasSuffix(assetName, ".tlz4") {
			return "tlz4"
		}
		return "tar.lz4"
	case *archiver.TarSz:
		if strings.HasSuffix(assetName, ".tsz") {
			return "tsz"
		}
		return "tar.sz"
	case *archiver.TarXz:
		if strings.HasSuffix(assetName, ".txz") {
			return "txz"
		}
		return "tar.xz"
	case *archiver.TarZstd:
		return "tar.zsd"
	case *archiver.Zip:
		return "zip"
	case *archiver.Gz:
		return "gz"
	case *archiver.Bz2:
		return "bz2"
	case *archiver.Lz4:
		return "lz4"
	case *archiver.Snappy:
		return "sz"
	case *archiver.Xz:
		return "xz"
	case *archiver.Zstd:
		return "zst"
	default:
		return formatRaw
	}
}
