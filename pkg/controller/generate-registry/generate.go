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
	"github.com/aquaproj/aqua/pkg/util"
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

func (ctrl *Controller) excludeAsset(pkgName, assetName, version string) bool {
	format := ctrl.getFormat(assetName)
	allowedExts := map[string]struct{}{
		".exe": {},
		".sh":  {},
		".js":  {},
		".jar": {},
		".py":  {},
	}
	if format == formatRaw {
		ext := util.Ext(assetName, version)
		if len(ext) > 0 && len(ext) < 6 {
			if _, ok := allowedExts[ext]; !ok {
				return true
			}
		}
	}
	suffixes := []string{
		"sha256",
	}
	asset := strings.ToLower(assetName)
	for _, s := range suffixes {
		if strings.HasSuffix(asset, "."+s) {
			return true
		}
	}
	words := []string{
		"changelog", "readme", "license", "openbsd", "freebsd", "netbsd", "android", "386", "i386", "armv6", "armv7", "32bit", "32-bit",
		"netbsd", "plan9", "solaris", "mips", "mips64", "mips64le", "mipsle", "ppc64", "ppc64le", "riscv64", "s390x", "wasm",
		"checksum",
	}
	for _, s := range words {
		if strings.Contains(asset, s) && !strings.Contains(pkgName, s) {
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
			logE.WithField("version", release.GetTagName()).Debug("got the latest release")
			assets := ctrl.listReleaseAssets(ctx, logE, pkgInfo, release.GetID())
			if len(assets) != 0 {
				logE.WithField("num_of_assets", len(assets)).Debug("got assets")
				assetInfos := make([]*AssetInfo, 0, len(assets))
				for _, asset := range assets {
					assetName := asset.GetName()
					if ctrl.excludeAsset(pkgName, assetName, release.GetTagName()) {
						logE.WithField("asset_name", assetName).Debug("exclude an asset")
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
	Score        int
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

func (ctrl *Controller) getOSArch(goos, goarch string, assetInfos []*AssetInfo) *AssetInfo { //nolint:gocognit,cyclop
	var a, rawA *AssetInfo
	for _, assetInfo := range assetInfos {
		assetInfo := assetInfo
		if (assetInfo.OS != goos || assetInfo.Arch != goarch) && !assetInfo.DarwinAll {
			continue
		}
		if assetInfo.Format == "" || assetInfo.Format == formatRaw { //nolint:nestif
			if rawA == nil {
				rawA = assetInfo
				continue
			}
			if assetInfo.Score > rawA.Score {
				rawA = assetInfo
				continue
			}
			rawAIdx := strings.Index(rawA.Template, "{")
			assetIdx := strings.Index(assetInfo.Template, "{")
			if rawAIdx != -1 && assetIdx != -1 {
				if rawAIdx > assetIdx {
					rawA = assetInfo
				}
				continue
			}
			if len(rawA.Template) > len(assetInfo.Template) {
				rawA = assetInfo
			}
			continue
		}
		if a == nil {
			a = assetInfo
			continue
		}
		if assetInfo.Score > a.Score {
			a = assetInfo
			continue
		}
		aIdx := strings.Index(a.Template, "{")
		assetIdx := strings.Index(assetInfo.Template, "{")
		if aIdx != -1 && assetIdx != -1 {
			if aIdx > assetIdx {
				a = assetInfo
			}
			continue
		}
		if len(a.Template) > len(assetInfo.Template) {
			a = assetInfo
		}
	}
	if a != nil {
		return a
	}
	return rawA
}

func mergeReplacements(m1, m2 map[string]string) (map[string]string, bool) {
	if len(m1) == 0 {
		return m2, true
	}
	if len(m2) == 0 {
		return m1, true
	}
	m := map[string]string{}
	for k, v1 := range m1 {
		m[k] = v1
		if v2, ok := m2[k]; ok && v1 != v2 {
			return nil, false
		}
	}
	for k, v2 := range m2 {
		if _, ok := m[k]; !ok {
			m[k] = v2
		}
	}
	return m, true
}

func (ctrl *Controller) parseAssetInfos(pkgInfo *registry.PackageInfo, assetInfos []*AssetInfo) { //nolint:funlen,gocognit,cyclop,gocyclo
	for _, goos := range []string{"linux", "darwin", "windows"} {
		var overrides []*registry.Override
		var supportedEnvs []string
		for _, goarch := range []string{"amd64", "arm64"} {
			if assetInfo := ctrl.getOSArch(goos, goarch, assetInfos); assetInfo != nil {
				overrides = append(overrides, &registry.Override{
					GOOS:         assetInfo.OS,
					GOArch:       assetInfo.Arch,
					Format:       assetInfo.Format,
					Replacements: assetInfo.Replacements,
					Asset:        strP(assetInfo.Template),
				})
				if goos == "darwin" && goarch == "amd64" {
					supportedEnvs = append(supportedEnvs, "darwin")
				} else {
					supportedEnvs = append(supportedEnvs, goos+"/"+goarch)
				}
			}
		}
		if len(overrides) == 2 { //nolint:gomnd
			supportedEnvs = []string{goos}
			asset1 := overrides[0]
			asset2 := overrides[1]
			if asset1.Format == asset2.Format && *asset1.Asset == *asset2.Asset {
				replacements, ok := mergeReplacements(overrides[0].Replacements, overrides[1].Replacements)
				if ok {
					overrides = []*registry.Override{
						{
							GOOS:         asset1.GOOS,
							Format:       asset1.Format,
							Replacements: replacements,
							Asset:        asset1.Asset,
						},
					}
				}
			}
		}
		if len(overrides) == 1 {
			overrides[0].GOArch = ""
		}
		pkgInfo.Overrides = append(pkgInfo.Overrides, overrides...)
		pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, supportedEnvs...)
	}

	darwinAmd64 := ctrl.getOSArch("darwin", "amd64", assetInfos)
	darwinArm64 := ctrl.getOSArch("darwin", "arm64", assetInfos)
	if darwinAmd64 != nil && darwinArm64 == nil {
		pkgInfo.Rosetta2 = boolP(true)
	}

	if reflect.DeepEqual(pkgInfo.SupportedEnvs, registry.SupportedEnvs{"linux", "darwin", "windows"}) {
		pkgInfo.SupportedEnvs = nil
	}
	if reflect.DeepEqual(pkgInfo.SupportedEnvs, registry.SupportedEnvs{"linux", "darwin", "windows/amd64"}) {
		pkgInfo.SupportedEnvs = registry.SupportedEnvs{"darwin", "linux", "amd64"}
	}
	if reflect.DeepEqual(pkgInfo.SupportedEnvs, registry.SupportedEnvs{"linux/amd64", "darwin", "windows/amd64"}) {
		pkgInfo.SupportedEnvs = registry.SupportedEnvs{"darwin", "amd64"}
	}

	formatCounts := map[string]int{}
	for _, override := range pkgInfo.Overrides {
		formatCounts[override.Format]++
	}
	maxCnt := 0
	for f, cnt := range formatCounts {
		if cnt > maxCnt {
			pkgInfo.Format = f
			maxCnt = cnt
			continue
		}
		if cnt == maxCnt && f != formatRaw {
			pkgInfo.Format = f
			maxCnt = cnt
			continue
		}
	}
	assetCounts := map[string]int{}
	for _, override := range pkgInfo.Overrides {
		override := override
		if override.Format != pkgInfo.Format {
			continue
		}
		override.Format = ""
		assetCounts[*override.Asset]++
	}
	maxCnt = 0
	for asset, cnt := range assetCounts {
		asset := asset
		if cnt > maxCnt {
			pkgInfo.Asset = &asset
			maxCnt = cnt
			continue
		}
	}
	overrides := []*registry.Override{}
	for _, override := range pkgInfo.Overrides {
		override := override
		if *override.Asset != *pkgInfo.Asset {
			overrides = append(overrides, override)
			continue
		}
		override.Asset = nil
		if override.Format != "" || len(override.Replacements) != 0 {
			overrides = append(overrides, override)
		}
	}
	pkgInfo.Overrides = overrides

	overrides = []*registry.Override{}
	for _, override := range pkgInfo.Overrides {
		override := override
		if len(override.Replacements) == 0 {
			overrides = append(overrides, override)
			continue
		}
		if pkgInfo.Replacements == nil {
			pkgInfo.Replacements = registry.Replacements{}
		}
		for k, v := range override.Replacements {
			vp, ok := pkgInfo.Replacements[k]
			if !ok {
				pkgInfo.Replacements[k] = v
				delete(override.Replacements, k)
				continue
			}
			if v == vp {
				delete(override.Replacements, k)
				continue
			}
		}
		if len(override.Replacements) != 0 || override.Format != "" || override.Asset != nil {
			overrides = append(overrides, override)
		}
	}
	pkgInfo.Overrides = overrides
	if len(pkgInfo.Overrides) == 0 && pkgInfo.Format != "" && pkgInfo.Format != formatRaw {
		asset := strings.Replace(*pkgInfo.Asset, "{{.Format}}", pkgInfo.Format, 1)
		pkgInfo.Asset = &asset
		pkgInfo.Format = ""
	}
}

const formatRaw = "raw"

func (ctrl *Controller) parseAssetName(assetName, version string) *AssetInfo { //nolint:cyclop
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
		if strings.Contains(lowAssetName, "_universal") || strings.Contains(lowAssetName, "-universal") || strings.Contains(lowAssetName, ".universal") {
			assetInfo.DarwinAll = true
		}
	}
	assetInfo.Format = ctrl.getFormat(assetName)
	if assetInfo.Format != formatRaw {
		assetInfo.Template = assetInfo.Template[:len(assetInfo.Template)-len(assetInfo.Format)] + "{{.Format}}"
	}
	if assetInfo.OS == "windows" && assetInfo.Format == formatRaw {
		assetInfo.Template = strings.TrimSuffix(assetInfo.Template, ".exe")
	}
	return assetInfo
}
