// Package checksumgetter obtains the checksum of every environment a package
// supports.
//
// Two things want these: "aqua update-checksum", which records them in
// aqua-checksums.json, and "aqua lock update", which records them in the lock file.
// They ask the same question of the same three sources, so they ask it here.
package checksumgetter

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/resolve"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// ChecksumFileVerifier checks the signature of a checksum file.
type ChecksumFileVerifier interface {
	VerifyChecksumFileContent(ctx context.Context, logger *slog.Logger, pkg *config.Package, assetName string, content []byte) error
}

// Getter reads checksums from the three sources.
type Getter struct {
	chkDL                download.ChecksumDownloader
	downloader           download.ClientAPI
	checksumFileVerifier ChecksumFileVerifier
}

func New(chkDL download.ChecksumDownloader, downloader download.ClientAPI, checksumFileVerifier ChecksumFileVerifier) *Getter {
	return &Getter{
		chkDL:                chkDL,
		downloader:           downloader,
		checksumFileVerifier: checksumFileVerifier,
	}
}

// Get records the checksum of every environment the package supports into checksums.
//
// Three sources answer, in this order. The GitHub Release API reports a digest for
// assets uploaded since 2025-06-03, which costs no download. A checksum file the
// package publishes is used when its signature can be verified, which covers one
// download for every environment. Otherwise the asset itself is downloaded and
// hashed, which always works and is what a package with no checksum configuration
// falls back to.
//
// An unsigned checksum file is never used: it is fetched from the same place as the
// asset, so believing it would add nothing to downloading the asset and hashing it.
func (g *Getter) Get(ctx context.Context, logger *slog.Logger, checksums *checksum.Checksums, pkg *config.Package, supportedEnvs []string) error {
	logger.Info("getting a package checksum")
	rts, err := checksum.GetRuntimesFromSupportedEnvs(supportedEnvs, pkg.PackageInfo.SupportedEnvs)
	if err != nil {
		return fmt.Errorf("get supported platforms: %w", err)
	}

	// Expand base runtimes by variant value combinations declared in the
	// package's Overrides. Without this, a package with multiple Overrides
	// for the same (GOOS, GOARCH) differing only by variants (e.g.
	// libc=musl vs libc=glibc) would have only one of its assets checksummed.
	rts = resolve.ExpandByVariants(pkg.PackageInfo, rts)

	pkgs, assetNames, err := g.getPkgs(pkg, rts)
	if err != nil {
		return err
	}

	// If every per-runtime checksum is already recorded, there is nothing to do.
	// This skips both the GitHub API pre-fetch below and the per-runtime loop.
	if allChecksumsCached(checksums, pkgs, rts) {
		return nil
	}

	releaseAssets := g.prefetchReleaseAssets(ctx, logger, pkgs, rts)

	checksumFiles := map[string]struct{}{}
	for _, rt := range rts {
		logger := logger.With(
			"checksum_env", rt.Env(),
		)
		if rt.LibC != "" {
			logger = logger.With("checksum_libc", rt.LibC)
		}
		pkg, ok := pkgs[resolve.RuntimeKey(rt)]
		if !ok {
			continue
		}
		if err := g.updatePackageByRuntime(ctx, logger, checksums, pkg, checksumFiles, rt, assetNames, releaseAssets); err != nil {
			return err
		}
	}
	return nil
}

func (g *Getter) getPkgs(pkg *config.Package, rts []*runtime.Runtime) (map[string]*config.Package, map[string]struct{}, error) {
	pkgs := make(map[string]*config.Package, len(rts))
	assets := make(map[string]struct{}, len(rts))
	for _, rt := range rts {
		pkgInfo := pkg.PackageInfo
		pkgInfo = pkgInfo.Copy()
		pkgInfo.OverrideByRuntime(rt)
		switch pkgInfo.Type {
		case "cargo", "go_install":
			// Skip updating checksums of these packages
			continue
		}
		pkgWithEnv := &config.Package{
			Package:     pkg.Package,
			PackageInfo: pkgInfo,
		}
		asset, err := pkgWithEnv.RenderAsset(rt)
		if err != nil {
			return nil, nil, fmt.Errorf("render an asset: %w", err)
		}
		assets[asset] = struct{}{}
		pkgs[resolve.RuntimeKey(rt)] = pkgWithEnv
	}
	return pkgs, assets, nil
}

// allChecksumsCached reports whether every per-runtime checksum for this
// package is already present in checksums. When true, the caller can skip
// all remaining work for this package (GitHub API pre-fetch, per-runtime
// download/verify, etc.).
//
// An empty checksum ID or an error from ChecksumID is treated as "not cached"
// so the regular path runs and the same error (if any) surfaces with full
// context in updatePackageByRuntime.
func allChecksumsCached(checksums *checksum.Checksums, pkgs map[string]*config.Package, rts []*runtime.Runtime) bool {
	for _, rt := range rts {
		p, ok := pkgs[resolve.RuntimeKey(rt)]
		if !ok {
			continue
		}
		id, err := p.ChecksumID(rt)
		if err != nil || id == "" {
			return false
		}
		if checksums.Get(id) == nil {
			return false
		}
	}
	return true
}

// prefetchReleaseAssets returns the GitHub Release Asset digests for pkgs if
// any per-runtime config qualifies (github_release type with no checksum-file
// signature verification). RepoOwner/RepoName/Version cannot be overridden per
// runtime, so one API call serves every matching runtime. Returns nil when no
// runtime can use the API or when the API call fails.
func (g *Getter) prefetchReleaseAssets(ctx context.Context, logger *slog.Logger, pkgs map[string]*config.Package, rts []*runtime.Runtime) domain.ReleaseAssets {
	for _, rt := range rts {
		p, ok := pkgs[resolve.RuntimeKey(rt)]
		if !ok {
			continue
		}
		if p.PackageInfo.Type != config.PkgInfoTypeGitHubRelease ||
			hasChecksumSignatureVerification(p.PackageInfo.Checksum) {
			continue
		}
		releaseAssets, err := g.chkDL.GetReleaseAssets(ctx, logger, p)
		if err != nil {
			slogerr.WithError(logger, err).Debug("failed to get release assets from GitHub API")
			return nil
		}
		return releaseAssets
	}
	return nil
}

// hasChecksumSignatureVerification returns true if the checksum has signature verification configured
// (Cosign, Minisign, or GitHubArtifactAttestations).
func hasChecksumSignatureVerification(chksum *registry.Checksum) bool {
	if chksum == nil {
		return false
	}
	return chksum.GetCosign() != nil || chksum.GetMinisign() != nil || chksum.GetGitHubArtifactAttestations() != nil
}

func (g *Getter) getChecksums(ctx context.Context, logger *slog.Logger, pkg *config.Package, checksumFiles map[string]struct{}, rt *runtime.Runtime, assetNames map[string]struct{}, checksumID string, releaseAssets domain.ReleaseAssets) ([]*checksum.Checksum, error) {
	if !pkg.PackageInfo.Checksum.GetEnabled() {
		cs, err := g.dlAssetAndGetChecksum(ctx, logger, pkg, rt, releaseAssets)
		if err != nil {
			return nil, err
		}
		return []*checksum.Checksum{cs}, nil
	}

	// If release assets were pre-fetched, try to get the digest from them.
	// Re-check the per-runtime config: another runtime may have triggered the
	// pre-fetch, but this runtime might have a per-runtime override that adds
	// signature verification or changes the type. Skipping this check would
	// bypass the signature verification of the checksum file.
	if releaseAssets != nil &&
		pkg.PackageInfo.Type == config.PkgInfoTypeGitHubRelease &&
		!hasChecksumSignatureVerification(pkg.PackageInfo.Checksum) {
		assetName, err := pkg.RenderAsset(rt)
		if err != nil {
			return nil, fmt.Errorf("render an asset: %w", err)
		}
		if digest := releaseAssets.GetDigest(assetName); digest != nil {
			logger.Debug("got digest from GitHub API",
				"checksum_id", checksumID,
				"checksum", digest.Digest)
			return []*checksum.Checksum{{
				ID:        checksumID,
				Checksum:  digest.Digest,
				Algorithm: digest.Algorithm,
			}}, nil
		}
	}

	checksumFileID, err := pkg.RenderChecksumFileID(rt)
	if err != nil {
		return nil, fmt.Errorf("render a checksum file ID: %w", err)
	}
	if _, ok := checksumFiles[checksumFileID]; ok {
		return nil, nil
	}
	checksumFiles[checksumFileID] = struct{}{}
	return g.dlAndVerifyChecksumFile(ctx, logger, pkg, rt, assetNames, checksumID)
}

func (g *Getter) dlAndVerifyChecksumFile(ctx context.Context, logger *slog.Logger, pkg *config.Package, rt *runtime.Runtime, assetNames map[string]struct{}, checksumID string) ([]*checksum.Checksum, error) {
	logger.Debug("downloading a checksum file")
	file, _, err := g.chkDL.DownloadChecksum(ctx, logger, rt, pkg)
	if err != nil {
		return nil, fmt.Errorf("download a checksum file: %w", err)
	}
	if file == nil {
		return nil, nil
	}
	defer file.Close()
	b, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read a checksum file: %w", err)
	}
	assetName, err := pkg.RenderAsset(rt)
	if err != nil {
		return nil, fmt.Errorf("render an asset name: %w", err)
	}
	if assetName != "" {
		assetName = filepath.Base(assetName)
	}
	if err := g.checksumFileVerifier.VerifyChecksumFileContent(ctx, logger, pkg, assetName, b); err != nil {
		return nil, fmt.Errorf("verify the checksum file: %w", err)
	}
	return g.getChecksumsFromChecksumFile(pkg, assetNames, checksumID, strings.TrimSpace(string(b)))
}

func (g *Getter) getChecksumsFromChecksumFile(pkg *config.Package, assetNames map[string]struct{}, checksumID string, checksumFile string) ([]*checksum.Checksum, error) {
	pkgInfo := pkg.PackageInfo
	m, s, err := checksum.ParseChecksumFile(checksumFile, pkgInfo.Checksum)
	if err != nil {
		return nil, fmt.Errorf("parse a checksum file: %w", err)
	}
	if s != "" {
		return []*checksum.Checksum{
			{
				ID:        checksumID,
				Checksum:  s,
				Algorithm: pkgInfo.Checksum.GetAlgorithm(),
			},
		}, nil
	}
	arr := make([]*checksum.Checksum, 0, len(m))
	for assetName, chksum := range m {
		if _, ok := assetNames[assetName]; !ok {
			continue
		}
		checksumID, err := pkg.ChecksumIDFromAsset(assetName)
		if err != nil {
			return nil, fmt.Errorf("get a checksum id from asset: %w", err)
		}
		arr = append(arr, &checksum.Checksum{
			ID:        checksumID,
			Checksum:  chksum,
			Algorithm: pkgInfo.Checksum.GetAlgorithm(),
		})
	}
	return arr, nil
}

func (g *Getter) updatePackageByRuntime(ctx context.Context, logger *slog.Logger, checksums *checksum.Checksums, pkg *config.Package, checksumFiles map[string]struct{}, rt *runtime.Runtime, assetNames map[string]struct{}, releaseAssets domain.ReleaseAssets) error {
	checksumID, err := pkg.ChecksumID(rt)
	if err != nil {
		return fmt.Errorf("get a checksum id: %w", err)
	}

	if a := checksums.Get(checksumID); a != nil {
		return nil
	}

	cs, err := g.getChecksums(ctx, logger, pkg, checksumFiles, rt, assetNames, checksumID, releaseAssets)
	if err != nil {
		return err
	}
	for _, c := range cs {
		if a := checksums.Get(c.ID); a != nil {
			continue
		}
		checksums.Set(c.ID, c)
	}
	return nil
}

func (g *Getter) dlAssetAndGetChecksum(ctx context.Context, logger *slog.Logger, pkg *config.Package, rt *runtime.Runtime, releaseAssets domain.ReleaseAssets) (*checksum.Checksum, error) {
	attrs := slogerr.NewAttrs(1)
	checksumID, err := pkg.ChecksumID(rt)
	if err != nil {
		return nil, fmt.Errorf("get a checksum id: %w", err)
	}
	assetName, err := pkg.RenderAsset(rt)
	if err != nil {
		return nil, fmt.Errorf("get an asset name: %w", err)
	}
	logger = attrs.Add(logger, "asset_name", assetName)

	// Try to get the digest from pre-fetched release assets.
	if releaseAssets != nil {
		if digest := releaseAssets.GetDigest(assetName); digest != nil {
			logger.Debug("got digest from GitHub API",
				"checksum_id", checksumID,
				"checksum", digest.Digest)
			return &checksum.Checksum{
				ID:        checksumID,
				Checksum:  digest.Digest,
				Algorithm: digest.Algorithm,
			}, nil
		}
	}

	logger.Info("downloading an asset to calculate the checksum")
	f, err := download.ConvertPackageToFile(pkg, assetName, rt)
	if err != nil {
		return nil, attrs.With(err) //nolint:wrapcheck
	}
	file, _, err := g.downloader.ReadCloser(ctx, logger, f)
	if err != nil {
		return nil, fmt.Errorf("download an asset: %w", attrs.With(err))
	}
	defer file.Close()
	algorithm := "sha256"
	chk, err := checksum.CalculateReader(file, algorithm)
	if err != nil {
		return nil, fmt.Errorf("calculate an asset: %w", slogerr.With(attrs.With(err), "algorithm", algorithm))
	}
	return &checksum.Checksum{
		ID:        checksumID,
		Checksum:  chk,
		Algorithm: algorithm,
	}, nil
}
