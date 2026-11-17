// Package lockresolve reads the definition of a locked package from its registry.
//
// The lock file records a package's checksums and command names, and nothing of where it
// is downloaded from. Where it comes from is always the registry's answer, read here: for
// the standard registry that is aqua-registry-g2, and for any other registry it is that
// registry's own definition.
//
// The lock update command uses this to build entries, and install and exec use it to
// install an entry.
package lockresolve

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/resolve"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

var (
	errRegistryNotFound      = errors.New("the registry isn't found")
	errPkgNotFoundInRegistry = errors.New("the package isn't found in the registry")
	// ErrNotInRegistry is an entry for an environment the registry has no build for.
	ErrNotInRegistry = errors.New("the registry has no build of the package for the environment the lock file has an entry for. Please run \"aqua lock update --force\"")
	// ErrChecksumMismatch is a registry and a lock file disagreeing on a checksum.
	ErrChecksumMismatch = errors.New("the checksum in the lock file doesn't match the one in the registry")
)

// G2 resolves a package of the standard registry from aqua-registry-g2.
type G2 interface {
	Resolve(ctx context.Context, logger *slog.Logger, pkgName, version string) ([]*lockfile.Package, error)
}

// RegistryInstaller installs one registry a configuration declares.
type RegistryInstaller interface {
	InstallRegistry(ctx context.Context, logger *slog.Logger, regist *aqua.Registry, cfgFilePath string, checksums *checksum.Checksums) (*registry.Config, error)
}

// Registries installs the registries of one configuration as the packages need them.
//
// A registry is installed only when a package from it is resolved, and only that one.
// A configuration whose packages all come from the standard registry therefore installs
// none of them, and a private registry is read only for a package that comes from it.
type Registries struct {
	installer   RegistryInstaller
	cfg         *aqua.Config
	cfgFilePath string
	contents    map[string]*registry.Config
}

// NewRegistries creates Registries for a configuration.
func NewRegistries(installer RegistryInstaller, cfg *aqua.Config, cfgFilePath string) *Registries {
	return &Registries{
		installer:   installer,
		cfg:         cfg,
		cfgFilePath: cfgFilePath,
		contents:    map[string]*registry.Config{},
	}
}

// Config is the configuration the registries are declared in.
func (r *Registries) Config() *aqua.Config {
	return r.cfg
}

// Get returns the content of the named registry, installing it the first time.
func (r *Registries) Get(ctx context.Context, logger *slog.Logger, name string) (*registry.Config, error) {
	if content, ok := r.contents[name]; ok {
		return content, nil
	}
	rgst, ok := r.cfg.Registries[name]
	if !ok || rgst == nil {
		return nil, errRegistryNotFound
	}
	// The registry file has a checksum of its own, which the installer verifies. A
	// fresh set is passed rather than aqua-checksums.json: the lock file is what this
	// path maintains, not that file.
	content, err := r.installer.InstallRegistry(ctx, logger, rgst, r.cfgFilePath, checksum.New())
	if err != nil {
		return nil, fmt.Errorf("install the registry: %w", slogerr.With(err, "registry_name", name))
	}
	r.contents[name] = content
	return content, nil
}

// Definitions is a package resolved for every environment it supports.
type Definitions struct {
	Packages []*lockfile.Package
	// PkgInfo and SupportedEnvs are what Packages were resolved from, for a package
	// from a registry other than the standard one. Its checksums aren't in the
	// registry, and fetching them needs these. Both are empty for the standard
	// registry, whose entries carry their checksums already.
	PkgInfo       *registry.PackageInfo
	SupportedEnvs []string
}

// Resolve resolves a package for every environment it supports.
//
// The standard registry is mirrored by aqua-registry-g2, which serves a registry.json
// resolved and checksummed ahead of time. Any other registry has no branch there, so its
// own definition is resolved here, and the checksums are left for the caller.
func Resolve(ctx context.Context, logger *slog.Logger, g2 G2, rgsts *Registries, pkg *aqua.Package) (*Definitions, error) {
	if isStandard(rgsts.Config().Registries[pkg.Registry]) {
		pkgs, err := g2.Resolve(ctx, logger, pkg.Name, pkg.Version)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
		return &Definitions{Packages: pkgs}, nil
	}
	return fromRegistry(ctx, logger, rgsts, pkg)
}

// isStandard says whether the registry is the standard one, which aqua-registry-g2
// mirrors.
//
// The name alone doesn't say it: a registry named "standard" can be declared with
// another repository, and its packages have no branch in aqua-registry-g2. Nor does
// "type: standard", which aqua.yaml turns into a github_content registry of
// aquaproj/aqua-registry when it is read. So the registry is the standard one when it is
// named "standard" and is aquaproj/aqua-registry. aquaproj/aqua-registry declared under
// another name is how a configuration keeps using aqua-registry v1, and it is resolved
// from that registry like any other.
func isStandard(rgst *aqua.Registry) bool {
	return rgst != nil &&
		rgst.Name == aqua.RegistryTypeStandard &&
		rgst.Type == aqua.RegistryTypeGitHubContent &&
		rgst.RepoOwner == "aquaproj" &&
		rgst.RepoName == "aqua-registry"
}

func fromRegistry(ctx context.Context, logger *slog.Logger, rgsts *Registries, pkg *aqua.Package) (*Definitions, error) {
	rgstContent, err := rgsts.Get(ctx, logger, pkg.Registry)
	if err != nil {
		return nil, err
	}
	pkgInfo := rgstContent.Package(logger, pkg.Name)
	if pkgInfo == nil {
		return nil, errPkgNotFoundInRegistry
	}

	cfg := rgsts.Config()
	var supportedEnvs []string
	if cfg.Checksum != nil {
		supportedEnvs = cfg.Checksum.SupportedEnvs
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

	// The checksums are fetched from the definition for this version, not from the
	// one the registry holds. A package that keeps everything in version_overrides
	// — suzuki-shunsuke/tfcmt has version_constraint "false" at the top and an
	// asset in every override — has no asset name at all until they are applied,
	// and the fetch asked for an empty one.
	versioned, err := pkgInfo.SetVersion(logger, pkg.Version)
	if err != nil {
		return nil, fmt.Errorf("apply version_overrides: %w", err)
	}

	rgst := cfg.Registries[pkg.Registry]
	for _, p := range pkgs {
		p.Registry = &lockfile.Registry{
			Type:      rgst.Type,
			RepoOwner: rgst.RepoOwner,
			RepoName:  rgst.RepoName,
		}
	}
	return &Definitions{
		Packages:      pkgs,
		PkgInfo:       versioned,
		SupportedEnvs: supportedEnvs,
	}, nil
}

// Locked returns the definition to install a lock file entry from.
//
// The definition is the registry's, for the entry's environment, and the checksum is the
// lock file's. Nothing else of the entry is used, so a lock file changed to name another
// download has nothing to name it with: what is downloaded is what the registry says,
// and it has to match the checksum the lock file recorded.
//
// A registry that carries checksums of its own, as aqua-registry-g2 does, is compared with
// the lock file before anything is downloaded. A difference means one of the two changed
// after the entry was written, and installing either would install something nobody
// locked.
func Locked(ctx context.Context, logger *slog.Logger, g2 G2, rgsts *Registries, pkg *aqua.Package, entry *lockfile.Entry) (*lockfile.Package, error) {
	defs, err := Resolve(ctx, logger, g2, rgsts, pkg)
	if err != nil {
		return nil, err
	}
	for _, def := range defs.Packages {
		if !entry.Describes(def) {
			continue
		}
		if err := setChecksum(def, entry); err != nil {
			return nil, err
		}
		return def, nil
	}
	return nil, ErrNotInRegistry
}

// setChecksum gives the definition the checksum the lock file recorded.
//
// A type that downloads nothing aqua could check, go_install and cargo, takes no checksum
// whatever the lock file says. Otherwise a checksum the registry has too must be the same
// one, algorithm included.
func setChecksum(def *lockfile.Package, entry *lockfile.Entry) error {
	if !def.NeedsChecksum() {
		def.Checksum = ""
		def.ChecksumAlgorithm = ""
		return nil
	}
	if def.Checksum != "" && (!strings.EqualFold(def.Checksum, entry.Checksum) || def.ChecksumAlgorithm != entry.ChecksumAlgorithm) {
		return slogerr.With(ErrChecksumMismatch, //nolint:wrapcheck
			"lock_file_checksum", entry.Checksum,
			"lock_file_checksum_algorithm", entry.ChecksumAlgorithm,
			"registry_checksum", def.Checksum,
			"registry_checksum_algorithm", def.ChecksumAlgorithm,
		)
	}
	def.Checksum = entry.Checksum
	def.ChecksumAlgorithm = entry.ChecksumAlgorithm
	return nil
}

// Source reads the definitions of a configuration's locked packages. It is Locked with
// its registries held.
type Source struct {
	g2    G2
	rgsts *Registries
}

// NewSource creates a Source.
func NewSource(g2 G2, rgsts *Registries) *Source {
	return &Source{g2: g2, rgsts: rgsts}
}

// Locked returns the definition to install a lock file entry from. See Locked.
func (s *Source) Locked(ctx context.Context, logger *slog.Logger, pkg *aqua.Package, entry *lockfile.Entry) (*lockfile.Package, error) {
	return Locked(ctx, logger, s.g2, s.rgsts, pkg, entry)
}
