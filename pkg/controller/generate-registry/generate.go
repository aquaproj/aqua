package genrgst

import (
	"context"
	"errors"
	"fmt"
	"regexp"
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
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

var errLimitMustBeGreaterEqualThanZero = errors.New("limit must be greater equal than zero")

func (c *Controller) GenerateRegistry(ctx context.Context, param *config.Param, logE *logrus.Entry, args ...string) error {
	if param.InitConfig {
		return c.initConfig(args...)
	}
	cfg := &Config{}
	if param.GenerateConfigFilePath != "" {
		if err := readConfig(c.fs, param.GenerateConfigFilePath, cfg); err != nil {
			return err
		}
	}

	args, err := parseArgs(args, cfg)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return nil
	}

	if param.Limit < 0 {
		return errLimitMustBeGreaterEqualThanZero
	}

	for _, arg := range args {
		if err := c.genRegistry(ctx, param, logE, cfg, arg); err != nil {
			return err
		}
	}
	return nil
}

func parseArgs(args []string, cfg *Config) ([]string, error) {
	if len(args) == 0 {
		if cfg.Package == "" {
			return nil, nil
		}
		return []string{cfg.Package}, nil
	}
	if cfg.Package != "" && cfg.Package != args[0] {
		return nil, logerr.WithFields(errors.New("a given package name is different from the package name in the configuration file"), logrus.Fields{ //nolint:wrapcheck
			"arg":               args[0],
			"package_in_config": cfg.Package,
		})
	}
	return args, nil
}

func (c *Controller) genRegistry(ctx context.Context, param *config.Param, logE *logrus.Entry, cfg *Config, pkgName string) error {
	pkgInfo, versions := c.getPackageInfo(ctx, logE, pkgName, param, cfg)
	if param.OutTestData != "" {
		if err := c.testdataOutputter.Output(&output.Param{
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
		encoder := yaml.NewEncoder(c.stdout, yaml.IndentSequence(true))
		if err := encoder.EncodeContext(ctx, cfg); err != nil {
			return fmt.Errorf("encode YAML: %w", err)
		}
		return nil
	}
	if err := c.insert(param.InsertFile, registry.PackageInfos{pkgInfo}); err != nil {
		return err
	}
	return nil
}

func (c *Controller) getRelease(ctx context.Context, repoOwner, repoName, version string) (*github.RepositoryRelease, error) {
	if version == "" {
		release, _, err := c.github.GetLatestRelease(ctx, repoOwner, repoName)
		return release, err //nolint:wrapcheck
	}
	release, _, err := c.github.GetReleaseByTag(ctx, repoOwner, repoName, version)
	return release, err //nolint:wrapcheck
}

func cleanDescription(desc string) string {
	return strings.TrimRight(strings.TrimSpace(gomoji.RemoveEmojis(desc)), ".!?")
}

func (c *Controller) getPackageInfo(ctx context.Context, logE *logrus.Entry, arg string, param *config.Param, cfg *Config) (*registry.PackageInfo, []string) {
	pkgInfo, versions := c.getPackageInfoMain(ctx, logE, arg, param, cfg)
	pkgInfo.Description = cleanDescription(pkgInfo.Description)
	if len(param.Commands) != 0 {
		files := make([]*registry.File, len(param.Commands))
		for i, cmd := range param.Commands {
			files[i] = &registry.File{
				Name: cmd,
			}
		}
		pkgInfo.Files = files
	}
	return pkgInfo, versions
}

func (c *Controller) getPackageInfoMain(ctx context.Context, logE *logrus.Entry, arg string, param *config.Param, cfg *Config) (*registry.PackageInfo, []string) { //nolint:cyclop
	pkgName, version, _ := strings.Cut(arg, "@")
	if strings.HasPrefix(pkgName, "crates.io/") {
		return c.getCargoPackageInfo(ctx, logE, pkgName)
	}
	splitPkgNames := strings.Split(pkgName, "/")
	pkgInfo := &registry.PackageInfo{
		Type:          "github_release",
		VersionPrefix: cfg.VersionPrefix,
	}
	if cfg.VersionFilter != nil {
		pkgInfo.VersionFilter = cfg.VersionFilter.Source().String()
	}
	if len(splitPkgNames) == 1 {
		pkgInfo.Name = pkgName
		return pkgInfo, nil
	}
	if len(splitPkgNames) != 2 { //nolint:mnd
		pkgInfo.Name = pkgName
	}
	pkgInfo.RepoOwner = splitPkgNames[0]
	pkgInfo.RepoName = splitPkgNames[1]
	repo, _, err := c.github.Get(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName)
	if err != nil {
		logE.WithFields(logrus.Fields{
			"repo_owner": pkgInfo.RepoOwner,
			"repo_name":  pkgInfo.RepoName,
		}).WithError(err).Warn("get the repository")
	} else {
		pkgInfo.Description = repo.GetDescription()
	}
	if param.Limit != 1 && version == "" {
		return pkgInfo, c.getPackageInfoWithVersionOverrides(ctx, logE, pkgName, pkgInfo, param, cfg)
	}
	release, err := c.getRelease(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, version)
	if err != nil {
		logE.WithFields(logrus.Fields{
			"repo_owner": pkgInfo.RepoOwner,
			"repo_name":  pkgInfo.RepoName,
		}).WithError(err).Warn("get the release")
		return pkgInfo, []string{version}
	}
	logE.WithField("version", release.GetTagName()).Debug("got the release")

	arr := c.listReleaseAssets(ctx, logE, pkgInfo, release.GetID())
	logE.WithField("num_of_assets", len(arr)).Debug("got assets")
	assetNames := make([]string, 0, len(arr))
	for _, asset := range arr {
		if excludeAsset(logE, asset.GetName(), cfg) {
			continue
		}
		assetNames = append(assetNames, asset.GetName())
	}

	c.patchRelease(logE, pkgInfo, pkgName, release.GetTagName(), assetNames)
	return pkgInfo, []string{version}
}

func getChecksum(checksumNames map[string]struct{}, assetName string) *registry.Checksum {
	suffixes := []string{
		"md5",
		"sha256",
		"sha512",
		"sha1",
	}
	for _, suffix := range suffixes {
		if _, ok := checksumNames[assetName+"."+suffix]; ok {
			return &registry.Checksum{
				Type:      "github_release",
				Asset:     "{{.Asset}}." + suffix,
				Algorithm: suffix,
			}
		}
	}
	return nil
}

func (c *Controller) patchRelease(logE *logrus.Entry, pkgInfo *registry.PackageInfo, pkgName, tagName string, assets []string) { //nolint:cyclop
	if len(assets) == 0 {
		pkgInfo.NoAsset = true
		return
	}
	assetInfos := make([]*asset.AssetInfo, 0, len(assets))
	pkgNameContainChecksum := strings.Contains(strings.ToLower(pkgName), "checksum")
	assetNames := map[string]struct{}{}
	checksumNames := map[string]struct{}{}
	for _, assetName := range assets {
		if !pkgNameContainChecksum {
			chksum := checksum.GetChecksumConfigFromFilename(assetName, tagName)
			if chksum != nil {
				checksumNames[assetName] = struct{}{}
				continue
			}
		}
		if asset.Exclude(pkgName, assetName) {
			logE.WithField("asset_name", assetName).Debug("exclude an asset")
			continue
		}
		assetNames[assetName] = struct{}{}
		assetInfo := asset.ParseAssetName(assetName, tagName)
		assetInfos = append(assetInfos, assetInfo)
	}
	for assetName := range assetNames {
		if checksum := getChecksum(checksumNames, assetName); checksum != nil {
			pkgInfo.Checksum = checksum
			break
		}
	}
	for assetName := range assetNames {
		if p := checkSLSAProvenance(assetName, tagName); p != nil {
			pkgInfo.SLSAProvenance = p
			break
		}
	}
	if len(checksumNames) > 0 && pkgInfo.Checksum == nil {
		for checksumName := range checksumNames {
			chksum := checksum.GetChecksumConfigFromFilename(checksumName, tagName)
			if chksum != nil {
				assetInfo := asset.ParseAssetName(checksumName, tagName)
				chksum.Asset = assetInfo.Template
				chksum.Cosign = checkChecksumCosign(pkgInfo, checksumName, assetNames)
				pkgInfo.Checksum = chksum
				break
			}
		}
	}
	asset.ParseAssetInfos(pkgInfo, assetInfos)
}

func (c *Controller) listReleaseAssets(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo, releaseID int64) []*github.ReleaseAsset {
	opts := &github.ListOptions{
		PerPage: 100, //nolint:mnd
	}
	var arr []*github.ReleaseAsset
	for range 10 {
		assets, _, err := c.github.ListReleaseAssets(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, releaseID, opts)
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

func checkSLSAProvenance(assetName, tagName string) *registry.SLSAProvenance {
	if !strings.HasSuffix(assetName, ".intoto.jsonl") {
		return nil
	}
	assetInfo := asset.ParseAssetName(assetName, tagName)
	return &registry.SLSAProvenance{
		Type:  "github_release",
		Asset: &assetInfo.Template,
	}
}

func findSignature(assetNames map[string]struct{}, checksumAssetName string) string {
	for _, suf := range []string{"-keyless.sig", ".sig"} {
		sig := checksumAssetName + suf
		if _, ok := assetNames[sig]; ok {
			return sig
		}
	}
	return ""
}

func findPubKey(assetNames map[string]struct{}) string {
	for assetName := range assetNames {
		if strings.HasSuffix(assetName, "cosign.pub") {
			return assetName
		}
	}
	return ""
}

func findCertificate(assetNames map[string]struct{}, checksumAssetName string) string {
	for _, suf := range []string{"-keyless.pem", ".pem"} {
		cert := checksumAssetName + suf
		if _, ok := assetNames[cert]; ok {
			return cert
		}
	}
	return ""
}

func findCosignBundle(assetNames map[string]struct{}, assetName string) string {
	for _, suf := range []string{".cosign.bundle", ".bundle", ".sigstore", ".sigstore.json"} {
		bundle := assetName + suf
		if _, ok := assetNames[bundle]; ok {
			return bundle
		}
	}
	return ""
}

func checkChecksumCosign(pkgInfo *registry.PackageInfo, checksumAssetName string, assetNames map[string]struct{}) *registry.Cosign { //nolint:cyclop
	cosign := &registry.Cosign{
		Opts: make([]string, 0, 8), //nolint:mnd // we generate max 8 arguments (certificate case)
	}
	downloadURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/{{.Version}}/",
		pkgInfo.RepoOwner, pkgInfo.RepoName)

	var bundleAssetName, certificateAssetName string
	if bundleAssetName = findCosignBundle(assetNames, checksumAssetName); bundleAssetName != "" {
		cosign.Bundle = &registry.DownloadedFile{
			Type:  "github_release",
			Asset: &bundleAssetName,
		}
	} else if certificateAssetName = findCertificate(assetNames, checksumAssetName); certificateAssetName != "" {
		cosign.Opts = append(cosign.Opts,
			"--certificate",
			downloadURL+certificateAssetName,
		)
	}
	if bundleAssetName != "" || certificateAssetName != "" {
		cosign.Opts = append(cosign.Opts,
			"--certificate-identity-regexp",
			fmt.Sprintf(
				`^https://github\.com/%s/%s/\.github/workflows/.+\.ya?ml@refs/tags/\Q{{.Version}}\E$`,
				regexp.QuoteMeta(pkgInfo.RepoOwner),
				regexp.QuoteMeta(pkgInfo.RepoName),
			),
			"--certificate-oidc-issuer",
			"https://token.actions.githubusercontent.com",
		)
	}

	// If a bundle was found, nothing else is needed
	if bundleAssetName != "" {
		return cosign
	}

	// For all other cases, signature is needed
	signatureAssetName := findSignature(assetNames, checksumAssetName)
	if signatureAssetName == "" {
		return nil
	}

	// If we do not have a certificate and the signature is not keyless, try public key
	if certificateAssetName == "" && !strings.HasSuffix(signatureAssetName, "-keyless.sig") {
		pubKeyAssetName := findPubKey(assetNames)
		if pubKeyAssetName != "" {
			cosign.Opts = append(cosign.Opts,
				"--key",
				downloadURL+pubKeyAssetName,
			)
		}
	}

	// Bail out if nothing we can use was found yet
	if len(cosign.Opts) == 0 {
		return nil
	}

	cosign.Opts = append(cosign.Opts,
		"--signature",
		downloadURL+signatureAssetName,
	)
	return cosign
}
