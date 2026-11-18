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
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

// fillChecksums records the checksum of each entry.
//
// The checksums are fetched for the whole package at once, because the sources answer
// that way: one GitHub API call reports the digest of every asset in a release, and
// one checksum file covers every environment. They come back keyed by the identifier
// aqua-checksums.json uses, which each entry can compute for itself.
func (c *Controller) fillChecksums(ctx context.Context, logger *slog.Logger, pkgs []*lockfile.Package, pkg *aqua.Package, pkgInfo *registry.PackageInfo, supportedEnvs []string) error {
	checksums := checksum.New()
	if err := c.checksumGetter.GetVerified(ctx, logger, checksums, &config.Package{
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
