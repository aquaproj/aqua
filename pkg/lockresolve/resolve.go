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
	"slices"
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
	errNoG2                  = errors.New("aqua-registry-g2 isn't available to read the package from")
	// ErrNotInRegistry is an entry for an environment the registry has no build for.
	ErrNotInRegistry = errors.New("the registry has no build of the package for the environment the lock file has an entry for. Please run \"aqua lock update --force\"")
	// ErrChecksumMismatch is a registry and a lock file disagreeing on a checksum.
	ErrChecksumMismatch = errors.New(`the checksum in the lock file doesn't match the one in the registry. If the registry is right, please run "aqua lock update --force"`)
	// ErrCommandsMismatch is a registry and a lock file disagreeing on the commands of a
	// package.
	ErrCommandsMismatch = errors.New(`the commands in the lock file don't match the ones in the registry. Please run "aqua lock update --force"`)
	// ErrRegistryMismatch is an entry resolved from another registry than the one
	// aqua.yaml now says the package comes from.
	ErrRegistryMismatch = errors.New(`the lock file entry was resolved from another registry than aqua.yaml declares. Please run "aqua lock update --force"`)
)

// Resolver resolves a package of the standard registry from aqua-registry-g2, for the
// lock update command. It says when the configuration asks for a package by a name the
// registry no longer uses.
type Resolver interface {
	Resolve(ctx context.Context, logger *slog.Logger, pkgName, version string) ([]*lockfile.Package, error)
}

// G2 reads the definitions of a package of the standard registry from aqua-registry-g2,
// for install and exec. It says nothing about renames: it runs on every command, and the
// lock update command is where that is said.
type G2 interface {
	Definitions(ctx context.Context, logger *slog.Logger, pkgName, version string) ([]*lockfile.Package, error)
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
	checksums   *checksum.Checksums
	contents    map[string]*registry.Config
}

// NewRegistries creates Registries for a configuration.
//
// checksums is aqua-checksums.json, which a registry file is checked against as it is
// without a lock file, or nil when it isn't enabled. The lock update command passes nil:
// the lock file is what it maintains.
func NewRegistries(installer RegistryInstaller, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) *Registries {
	return &Registries{
		installer:   installer,
		cfg:         cfg,
		cfgFilePath: cfgFilePath,
		checksums:   checksums,
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
	// The registry file has a checksum of its own, which the installer checks against
	// the set it is given.
	chks := r.checksums
	if chks == nil {
		chks = checksum.New()
	}
	content, err := r.installer.InstallRegistry(ctx, logger, rgst, r.cfgFilePath, chks)
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
func Resolve(ctx context.Context, logger *slog.Logger, g2 Resolver, rgsts *Registries, pkg *aqua.Package) (*Definitions, error) {
	var fetch fetchFunc
	if g2 != nil {
		fetch = g2.Resolve
	}
	return resolveWith(ctx, logger, fetch, rgsts, pkg)
}

type fetchFunc func(ctx context.Context, logger *slog.Logger, pkgName, version string) ([]*lockfile.Package, error)

func resolveWith(ctx context.Context, logger *slog.Logger, fetch fetchFunc, rgsts *Registries, pkg *aqua.Package) (*Definitions, error) {
	if IsStandard(rgsts.Config().Registries[pkg.Registry]) {
		if fetch == nil {
			return nil, errNoG2
		}
		pkgs, err := fetch(ctx, logger, pkg.Name, pkg.Version)
		if err != nil {
			return nil, err
		}
		return &Definitions{Packages: pkgs}, nil
	}
	return fromRegistry(ctx, logger, rgsts, pkg)
}

// IsStandard says whether the registry is the standard one, which aqua-registry-g2
// mirrors.
//
// The name alone doesn't say it: a registry named "standard" can be declared with
// another repository, and its packages have no branch in aqua-registry-g2. Nor does
// "type: standard", which aqua.yaml turns into a github_content registry of
// aquaproj/aqua-registry when it is read. So the registry is the standard one when it is
// named "standard" and is registry.yaml of aquaproj/aqua-registry, the file
// aqua-registry-g2 mirrors. aquaproj/aqua-registry declared under another name is how a
// configuration keeps using aqua-registry v1, and it is resolved from that registry like
// any other.
func IsStandard(rgst *aqua.Registry) bool {
	return rgst != nil &&
		rgst.Name == aqua.RegistryTypeStandard &&
		rgst.Type == aqua.RegistryTypeGitHubContent &&
		rgst.RepoOwner == "aquaproj" &&
		rgst.RepoName == "aqua-registry" &&
		rgst.Path == "registry.yaml"
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
// What the lock file does record is compared with the registry before anything is
// downloaded: the registry the entry was resolved from, the commands, and a checksum the
// registry carries too, as aqua-registry-g2 does. A difference means one of the two
// changed after the entry was written, and installing either would install something
// nobody locked. The commands matter on their own: links are created from the lock
// file's, and a command the registry's definition doesn't have would be a link that
// shadows a command of the system.
func Locked(ctx context.Context, logger *slog.Logger, g2 G2, rgsts *Registries, pkg *aqua.Package, entry *lockfile.Entry) (*lockfile.Package, error) {
	var fetch fetchFunc
	if g2 != nil {
		fetch = g2.Definitions
	}
	defs, err := resolveWith(ctx, logger, fetch, rgsts, pkg)
	if err != nil {
		return nil, err
	}
	for _, def := range defs.Packages {
		if !entry.Describes(def) {
			continue
		}
		if !sameRegistry(entry.Registry, def.Registry) {
			return nil, ErrRegistryMismatch
		}
		if got := def.Entry().Commands; !slices.Equal(got, entry.Commands) {
			return nil, slogerr.With(ErrCommandsMismatch, //nolint:wrapcheck
				"lock_file_commands", strings.Join(entry.Commands, ","),
				"registry_commands", strings.Join(got, ","),
			)
		}
		if err := setChecksum(def, entry); err != nil {
			return nil, err
		}
		return def, nil
	}
	return nil, ErrNotInRegistry
}

// sameRegistry says whether an entry was resolved from the registry a definition comes
// from. Only the repository is compared, which is what the entry records.
func sameRegistry(a, b *lockfile.Registry) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Type == b.Type && a.RepoOwner == b.RepoOwner && a.RepoName == b.RepoName
}

// setChecksum gives the definition the checksum the lock file recorded.
//
// A type that downloads nothing aqua could check, go_install and cargo, takes no checksum,
// and an entry locked with one is one the registry no longer describes: dropping the
// checksum would let a registry changed to build from source skip the check the lock file
// recorded. Otherwise a checksum the registry has too must be the same one, algorithm
// included.
func setChecksum(def *lockfile.Package, entry *lockfile.Entry) error {
	if !def.NeedsChecksum() {
		if entry.Checksum != "" {
			return slogerr.With(ErrChecksumMismatch, //nolint:wrapcheck
				"lock_file_checksum", entry.Checksum,
				"registry_type", def.Type,
			)
		}
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
