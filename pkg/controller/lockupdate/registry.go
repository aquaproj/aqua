package lockupdate

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/resolve"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

// registries installs the registries a configuration declares, once.
//
// Installing is deferred until a package needs one, so a configuration whose packages
// all come from aqua-registry-g2 downloads none of them.
type registries struct {
	installer   RegistryInstaller
	cfg         *aqua.Config
	cfgFilePath string
	contents    map[string]*registry.Config
}

func (r *registries) Get(ctx context.Context, logger *slog.Logger, name string) (*registry.Config, error) {
	if r.contents == nil {
		// The registry file has a checksum of its own, which the installer
		// verifies. A fresh set is passed rather than aqua-checksums.json, because
		// this command's output is the lock file and it doesn't maintain that file.
		contents, err := r.installer.InstallRegistries(ctx, logger, r.cfg, r.cfgFilePath, checksum.New())
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
		r.contents = contents
	}
	rgst, ok := r.contents[name]
	if !ok {
		return nil, errRegistryNotFound
	}
	return rgst, nil
}

// resolveFromRegistry builds the lock file entries of a package from the registry it
// is declared in.
//
// This is the arm for registries other than the standard one, which have no branch in
// aqua-registry-g2 to read a generated registry.json from. The registry's own
// definition is resolved for every environment and the checksums are fetched, which
// is the work ar2 does ahead of time for the packages g2 covers.
func (c *Controller) resolveFromRegistry(ctx context.Context, logger *slog.Logger, rgsts *registries, pkg *aqua.Package) ([]*lockfile.Package, error) {
	rgstContent, err := rgsts.Get(ctx, logger, pkg.Registry)
	if err != nil {
		return nil, err
	}
	pkgInfo := rgstContent.Package(logger, pkg.Name)
	if pkgInfo == nil {
		return nil, errPkgNotFoundInRegistry
	}

	var supportedEnvs []string
	if rgsts.cfg.Checksum != nil {
		supportedEnvs = rgsts.cfg.Checksum.SupportedEnvs
	}
	pkgs, err := resolve.Resolve(logger, &resolve.Param{
		PkgName:       pkg.Name,
		Version:       pkg.Version,
		PkgInfo:       pkgInfo,
		SupportedEnvs: supportedEnvs,
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	if err := c.fillChecksums(ctx, logger, pkgs, pkg, pkgInfo, supportedEnvs); err != nil {
		return nil, err
	}

	rgst := rgsts.cfg.Registries[pkg.Registry]
	for _, p := range pkgs {
		p.Registry = &lockfile.Registry{
			Type:      rgst.Type,
			RepoOwner: rgst.RepoOwner,
			RepoName:  rgst.RepoName,
		}
	}
	return pkgs, nil
}

// fillChecksums records the checksum of each entry.
//
// The checksums are fetched for the whole package at once, because the sources answer
// that way: one GitHub API call reports the digest of every asset in a release, and
// one checksum file covers every environment. They come back keyed by the identifier
// aqua-checksums.json uses, which each entry can compute for itself.
func (c *Controller) fillChecksums(ctx context.Context, logger *slog.Logger, pkgs []*lockfile.Package, pkg *aqua.Package, pkgInfo *registry.PackageInfo, supportedEnvs []string) error {
	checksums := checksum.New()
	if err := c.checksumGetter.Get(ctx, logger, checksums, &config.Package{
		Package:     pkg,
		PackageInfo: pkgInfo,
	}, supportedEnvs); err != nil {
		return err //nolint:wrapcheck
	}

	for _, p := range pkgs {
		id, err := checksumID(p)
		if err != nil {
			return err
		}
		chksum := checksums.Get(id)
		if chksum == nil {
			// An entry whose type downloads an artifact has to carry a checksum,
			// or the lock file stops being what aqua verifies against.
			if p.NeedsChecksum() {
				return errNoChecksum
			}
			continue
		}
		// aqua-checksums.json holds checksums uppercased, while what ar2 writes into
		// registry.json is the lowercase hex GitHub reports. Both verify the same,
		// since the comparison ignores case, but a lock file mixing the two would
		// read as though the entries came from different things. They are written
		// the way the generated ones are.
		p.Checksum = strings.ToLower(chksum.Checksum)
		p.ChecksumAlgorithm = chksum.Algorithm
	}
	return nil
}

// checksumID is the identifier the entry's checksum was recorded under. The entry is
// already resolved, so turning it back into a package renders no templates and the
// identifier is computed by the same code that recorded it.
func checksumID(p *lockfile.Package) (string, error) {
	pkg := &config.Package{
		Package:     &aqua.Package{Name: p.Name, Version: p.Version},
		PackageInfo: p.PackageInfo(),
	}
	id, err := pkg.ChecksumID(&runtime.Runtime{
		GOOS:   p.OS,
		GOARCH: p.Arch,
		LibC:   p.Variants["libc"],
	})
	if err != nil {
		return "", fmt.Errorf("get a checksum id: %w", err)
	}
	return id, nil
}
