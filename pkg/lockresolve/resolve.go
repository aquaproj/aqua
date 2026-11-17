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
	"os"
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
	errNoRegistryChecksum    = errors.New("the registry file wasn't downloaded, so its checksum can't be recorded")
	errPkgNotFoundInRegistry = errors.New("the package isn't found in the registry")
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
	// ErrRegistryNotInLockFile is a registry file the lock file has no checksum of.
	ErrRegistryNotInLockFile = errors.New(`the lock file has no checksum of the registry. Please run "aqua lock update"`)
	// ErrRegistryChecksumMismatch is a registry file whose checksum isn't the one the
	// lock file recorded.
	ErrRegistryChecksumMismatch = errors.New(`the checksum of the registry doesn't match the one in the lock file. If the registry is right, please run "aqua lock update --force"`)
	// ErrWeakRegistryChecksum is a registry checksum of an algorithm that isn't
	// collision resistant.
	ErrWeakRegistryChecksum = errors.New(`the checksum of the registry in the lock file must be sha256 or sha512. Please run "aqua lock update --force"`)
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
//
// The registry decides where a package is downloaded from, so a github_content registry
// file is checked against the checksum the lock file recorded for it. Install and exec
// require one; the lock update command records it.
type Registries struct {
	installer   RegistryInstaller
	cfg         *aqua.Config
	cfgFilePath string
	lf          *lockfile.LockFile
	// rootDir is set when recording, to read a registry file already on disk: the
	// installer checks a file only when it downloads one.
	rootDir string
	record  bool
	force   bool
	// pending holds the checksums recorded this run, by registry name, until Commit
	// writes them to the lock file.
	pending  map[string]*lockfile.Registry
	contents map[string]*registry.Config
}

// NewRegistries creates Registries that check every github_content registry file against
// the checksum lf recorded, for install and exec.
func NewRegistries(installer RegistryInstaller, cfg *aqua.Config, cfgFilePath string, lf *lockfile.LockFile) *Registries {
	return &Registries{
		installer:   installer,
		cfg:         cfg,
		cfgFilePath: cfgFilePath,
		lf:          lf,
		contents:    map[string]*registry.Config{},
	}
}

// NewRecordingRegistries creates Registries that record the checksum of every
// github_content registry file in lf, for the lock update command. A checksum lf already
// has is checked rather than replaced, unless force is set: then the file is downloaded
// again and its checksum replaces the recorded one once it has been read.
func NewRecordingRegistries(installer RegistryInstaller, cfg *aqua.Config, cfgFilePath string, lf *lockfile.LockFile, rootDir string, force bool) *Registries {
	r := NewRegistries(installer, cfg, cfgFilePath, lf)
	r.record = true
	r.rootDir = rootDir
	r.force = force
	r.pending = map[string]*lockfile.Registry{}
	return r
}

// Config is the configuration the registries are declared in.
func (r *Registries) Config() *aqua.Config {
	return r.cfg
}

// Commit writes the checksums recorded this run to the lock file, for the registries ok
// accepts, and says whether the lock file changed.
//
// The caller leaves out a registry a package failed to be resolved from. A checksum
// recorded for it would sit beside entries resolved from the file it replaced, and the
// next run would see the registry recorded and leave those entries alone.
func (r *Registries) Commit(ok func(name string) bool) bool {
	changed := false
	for name, key := range r.pending {
		if !ok(name) {
			continue
		}
		// Recording again what the lock file already says, as --force does, isn't
		// a change.
		if old := r.lf.FindRegistry(key); old != nil && old.Checksum == key.Checksum && old.ChecksumAlgorithm == key.ChecksumAlgorithm {
			continue
		}
		r.lf.SetRegistry(key)
		changed = true
	}
	return changed
}

// LockFileRegistry is what the lock file records of a registry file, without the checksum.
// It is nil for a registry that has no checksum in the lock file: a local one, which sits
// in the repository beside it.
func LockFileRegistry(rgst *aqua.Registry) *lockfile.Registry {
	if rgst == nil || rgst.Type != aqua.RegistryTypeGitHubContent {
		return nil
	}
	return &lockfile.Registry{
		Type:      rgst.Type,
		RepoOwner: rgst.RepoOwner,
		RepoName:  rgst.RepoName,
		Ref:       rgst.Ref,
		Path:      rgst.Path,
	}
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
	content, err := r.install(ctx, logger, rgst)
	if err != nil {
		return nil, fmt.Errorf("install the registry: %w", slogerr.With(err, "registry_name", name))
	}
	r.contents[name] = content
	return content, nil
}

func (r *Registries) install(ctx context.Context, logger *slog.Logger, rgst *aqua.Registry) (*registry.Config, error) {
	key := LockFileRegistry(rgst)
	if key == nil {
		return r.installer.InstallRegistry(ctx, logger, rgst, r.cfgFilePath, checksum.New()) //nolint:wrapcheck
	}
	recorded, restore, err := r.recorded(rgst, key)
	if err != nil {
		return nil, err
	}
	// The installer compares a file it downloads with the checksum in the set, and
	// records one when the set has none. aqua-checksums.json isn't consulted: the lock
	// file has taken its place.
	chks := checksum.New()
	id := checksum.RegistryID(rgst)
	if recorded != nil {
		chks.Set(id, &checksum.Checksum{ID: id, Checksum: recorded.Checksum, Algorithm: recorded.ChecksumAlgorithm})
	}
	content, err := r.installer.InstallRegistry(ctx, logger, rgst, r.cfgFilePath, chks)
	restore(err != nil)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	if r.record {
		if err := r.recordChecksum(rgst, key, recorded, chks.Get(id)); err != nil {
			return nil, err
		}
	}
	return content, nil
}

// recorded is the checksum the registry file is checked against, or nil when it is to be
// recorded instead. restore is to be called once the file is installed, with whether that
// failed.
func (r *Registries) recorded(rgst *aqua.Registry, key *lockfile.Registry) (*lockfile.Registry, func(failed bool), error) {
	noop := func(bool) {}
	recorded := r.lf.FindRegistry(key)
	if recorded != nil && !strongAlgorithm(recorded.ChecksumAlgorithm) && (!r.record || !r.force) {
		return nil, noop, slogerr.With(ErrWeakRegistryChecksum, "checksum_algorithm", recorded.ChecksumAlgorithm) //nolint:wrapcheck
	}
	if !r.record {
		if recorded == nil {
			return nil, noop, ErrRegistryNotInLockFile
		}
		return recorded, noop, nil
	}
	if recorded != nil && !r.force {
		return recorded, noop, nil
	}
	// A checksum is recorded only from a file just downloaded. One already on disk may
	// be stale or edited, and a checksum of it would fail on every machine that
	// downloads the real one.
	restore, err := r.setAsideInstalledFile(rgst)
	if err != nil {
		return nil, noop, err
	}
	return nil, restore, nil
}

// strongAlgorithm says whether a checksum of the algorithm can stand for a registry file.
func strongAlgorithm(algorithm string) bool {
	return algorithm == "sha256" || algorithm == "sha512"
}

// setAsideInstalledFile moves the registry file installed before, and the JSON made from
// it, out of the way so that the installer downloads it again. The returned function puts
// them back if the download failed, and removes them otherwise: the registry file is
// shared by every configuration using the same registry at the same ref, and a failed
// download, offline for instance, mustn't cost them their copy.
func (r *Registries) setAsideInstalledFile(rgst *aqua.Registry) (func(failed bool), error) {
	filePath, err := rgst.FilePath(r.rootDir, r.cfgFilePath)
	if err != nil {
		return nil, fmt.Errorf("get the path of the registry file: %w", err)
	}
	var moved []string
	for _, p := range []string{filePath, filePath + ".json"} {
		if err := os.Rename(p, p+setAsideSuffix); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("set the installed registry file aside: %w", err)
		}
		moved = append(moved, p)
	}
	return func(failed bool) {
		for _, p := range moved {
			if failed {
				_ = os.Rename(p+setAsideSuffix, p)
				continue
			}
			_ = os.Remove(p + setAsideSuffix)
		}
	}, nil
}

// setAsideSuffix is appended to a registry file while it is downloaded again.
const setAsideSuffix = ".aqua-lock-update"

// recordChecksum records the checksum of the registry file on disk, or checks it against
// the one recorded. The file is read rather than taken from the installer, which computes
// one only for a file it downloads.
//
// A file just downloaded has been checksummed by the installer, which recorded it in
// downloaded, and that is the checksum recorded. A file already on disk, which the
// installer didn't read, is checksummed here to be checked against the recorded one.
func (r *Registries) recordChecksum(rgst *aqua.Registry, key, recorded *lockfile.Registry, downloaded *checksum.Checksum) error {
	if recorded == nil {
		if downloaded == nil {
			return errNoRegistryChecksum
		}
		key.Checksum = strings.ToLower(downloaded.Checksum)
		key.ChecksumAlgorithm = downloaded.Algorithm
		r.pending[rgst.Name] = key
		return nil
	}
	algorithm := "sha256"
	if strongAlgorithm(recorded.ChecksumAlgorithm) {
		algorithm = recorded.ChecksumAlgorithm
	}
	sum, err := r.fileChecksum(rgst, algorithm)
	if err != nil {
		return err
	}
	if !strings.EqualFold(recorded.Checksum, sum) {
		return slogerr.With(ErrRegistryChecksumMismatch, //nolint:wrapcheck
			"lock_file_checksum", recorded.Checksum,
			"registry_checksum", sum,
		)
	}
	return nil
}

// fileChecksum is the checksum of the registry file installed on disk, in lowercase.
func (r *Registries) fileChecksum(rgst *aqua.Registry, algorithm string) (string, error) {
	filePath, err := rgst.FilePath(r.rootDir, r.cfgFilePath)
	if err != nil {
		return "", fmt.Errorf("get the path of the registry file: %w", err)
	}
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open the registry file: %w", err)
	}
	defer f.Close()
	sum, err := checksum.CalculateReader(f, algorithm)
	if err != nil {
		return "", fmt.Errorf("calculate the checksum of the registry file: %w", err)
	}
	return strings.ToLower(sum), nil
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
	return resolveWith(ctx, logger, g2.Resolve, rgsts, pkg)
}

type fetchFunc func(ctx context.Context, logger *slog.Logger, pkgName, version string) ([]*lockfile.Package, error)

func resolveWith(ctx context.Context, logger *slog.Logger, fetch fetchFunc, rgsts *Registries, pkg *aqua.Package) (*Definitions, error) {
	if IsStandard(rgsts.Config().Registries[pkg.Registry]) {
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
// named "standard" and is aquaproj/aqua-registry. aquaproj/aqua-registry declared under
// another name is how a configuration keeps using aqua-registry v1, and it is resolved
// from that registry like any other.
func IsStandard(rgst *aqua.Registry) bool {
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
// What the lock file does record is compared with the registry before anything is
// downloaded: the registry the entry was resolved from, the commands, and a checksum the
// registry carries too, as aqua-registry-g2 does. A difference means one of the two
// changed after the entry was written, and installing either would install something
// nobody locked. The commands matter on their own: links are created from the lock
// file's, and a command the registry's definition doesn't have would be a link that
// shadows a command of the system.
func Locked(ctx context.Context, logger *slog.Logger, g2 G2, rgsts *Registries, pkg *aqua.Package, entry *lockfile.Entry) (*lockfile.Package, error) {
	defs, err := resolveWith(ctx, logger, g2.Definitions, rgsts, pkg)
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
// from. Only the repository is compared: the ref of a registry the lock file reads is
// checked through the registry file's checksum.
func sameRegistry(a, b *lockfile.Registry) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Type == b.Type && a.RepoOwner == b.RepoOwner && a.RepoName == b.RepoName
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
