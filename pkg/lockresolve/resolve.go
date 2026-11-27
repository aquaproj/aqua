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
	"sync"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/g2"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/resolve"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

var (
	errRegistryNotFound      = errors.New("the registry isn't found")
	errNoRegistryChecksum    = errors.New("the checksum of the downloaded registry file wasn't recorded")
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
	// ErrRegistryNotInLockFile is a registry file the lock file has no commit and checksum
	// of, which a lock file written before the commit was recorded lacks too.
	ErrRegistryNotInLockFile = errors.New(`the lock file has no commit or checksum of the registry. Please run "aqua lock update"`)
	// ErrWeakRegistryChecksum is a registry checksum of an algorithm that isn't
	// collision resistant.
	ErrWeakRegistryChecksum = errors.New(`the checksum of the registry in the lock file must be sha256 or sha512. Please run "aqua lock update --force"`)
)

// commitUnavailableHint is logged with a registry that couldn't be read at the recorded
// commit.
const commitUnavailableHint = `if the commit of the registry in the lock file is no longer available, run "aqua lock update --force" to record the commit the ref points at now`

// Resolver resolves a package of the standard registry from aqua-registry-g2, for the
// lock update command. It says when the configuration asks for a package by a name the
// registry no longer uses.
type Resolver interface {
	Resolve(ctx context.Context, logger *slog.Logger, pkgName, version string, supportedEnvs []string) (*g2.Resolved, error)
}

// G2 reads the definitions of a package of the standard registry from aqua-registry-g2,
// for install and exec. It says nothing about renames: it runs on every command, and the
// lock update command is where that is said.
type G2 interface {
	Definitions(ctx context.Context, logger *slog.Logger, pkgName, version string, supportedEnvs []string) (*g2.Resolved, error)
}

// RegistryInstaller installs one registry a configuration declares. It uses the registry
// file installed before when there is one, and checks a file against the checksums only
// when it downloads one.
//
// InstallLockedRegistry reads a github_content registry at the commit the lock file
// recorded, downloading it from the commit and checking it against the checksum when it
// isn't installed yet.
type RegistryInstaller interface {
	InstallRegistry(ctx context.Context, logger *slog.Logger, regist *aqua.Registry, cfgFilePath string, checksums *checksum.Checksums) (*registry.Config, error)
	InstallLockedRegistry(ctx context.Context, logger *slog.Logger, regist *aqua.Registry, commit string, chk *checksum.Checksum) (*registry.Config, error)
}

// RegistryDownloader is a RegistryInstaller that can also download a registry whatever is
// installed, and check it against the checksums, which is what recording a checksum needs.
type RegistryDownloader interface {
	RegistryInstaller
	DownloadRegistry(ctx context.Context, logger *slog.Logger, regist *aqua.Registry, commit string, checksums *checksum.Checksums) (*registry.Config, error)
}

// CommitResolver resolves a registry's ref to the commit it points at now, which the lock
// update command records.
type CommitResolver interface {
	GetCommitSHA1(ctx context.Context, owner, repo, ref, lastSHA string) (string, *github.Response, error)
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
//
// It is safe to use from more than one goroutine: install resolves its packages in
// parallel. Each registry is installed once, and what came of it, failure included, is
// what every package from it gets.
type Registries struct {
	installer   RegistryInstaller
	downloader  RegistryDownloader
	resolver    CommitResolver
	cfg         *aqua.Config
	cfgFilePath string
	lf          *lockfile.LockFile
	record      bool
	force       func(key *lockfile.Registry) bool
	// readRef makes a github_content registry the lock file has no record of read at its
	// ref, checked against refChecksums, rather than fail. The update command does that.
	readRef      bool
	refChecksums *checksum.Checksums
	// mu guards pending and installs.
	mu sync.Mutex
	// pending holds the checksums recorded this run, by registry name, until Commit
	// writes them to the lock file.
	pending map[string]*lockfile.Registry
	// installs holds what installing each registry came to, by registry name.
	installs map[string]*registryInstall
}

// registryInstall is the one installation of a registry, and what came of it.
type registryInstall struct {
	once    sync.Once
	content *registry.Config
	err     error
}

// NewRegistries creates Registries that check every github_content registry file against
// the checksum lf recorded, for install and exec.
func NewRegistries(installer RegistryInstaller, cfg *aqua.Config, cfgFilePath string, lf *lockfile.LockFile) *Registries {
	return &Registries{
		installer:   installer,
		cfg:         cfg,
		cfgFilePath: cfgFilePath,
		lf:          lf,
		installs:    map[string]*registryInstall{},
	}
}

// NewRecordingRegistries creates Registries that record the commit and the checksum of every
// github_content registry file in lf, for the lock update command. Every such registry is
// downloaded, whatever is installed: the checksum recorded has to be the checksum of what
// the commit holds.
//
// A registry lf already has a commit for is downloaded from that commit and checked against
// the recorded checksum, so a ref that has moved on neither fails the run nor changes the
// lock file. A registry without one, or one force says so for, has its ref resolved to the
// commit it points at now, and that commit and its checksum are recorded.
func NewRecordingRegistries(downloader RegistryDownloader, resolver CommitResolver, cfg *aqua.Config, cfgFilePath string, lf *lockfile.LockFile, force func(key *lockfile.Registry) bool) *Registries {
	r := NewRegistries(downloader, cfg, cfgFilePath, lf)
	r.downloader = downloader
	r.resolver = resolver
	r.record = true
	r.force = force
	r.pending = map[string]*lockfile.Registry{}
	return r
}

// NewUpdateRegistries creates Registries for the update command: a github_content registry
// the lock file has a record of is read at the recorded commit and checked against the
// recorded checksum, as install does, and one it has none of is read at its ref, as it was
// before there were lock files, and checked against checksums, aqua-checksums.json when
// it is enabled.
//
// A registry with no record is what a project has after adding a registry and before the
// next lock update, and after "aqua update" moved the registry's ref. All the update
// command needs from it is how the versions of a package are found, and failing because
// the lock file is behind would only make "aqua update" depend on "aqua lock update" for
// no gain: nothing is installed from what is read.
func NewUpdateRegistries(installer RegistryInstaller, cfg *aqua.Config, cfgFilePath string, lf *lockfile.LockFile, checksums *checksum.Checksums) *Registries {
	r := NewRegistries(installer, cfg, cfgFilePath, lf)
	r.readRef = true
	r.refChecksums = checksums
	return r
}

// Config is the configuration the registries are declared in.
func (r *Registries) Config() *aqua.Config {
	return r.cfg
}

// Commit writes the checksums recorded this run to the lock file, for the registry files ok
// accepts, and says whether the lock file changed.
//
// The caller leaves out a registry a package failed to be resolved from. A checksum
// recorded for it would sit beside entries resolved from the file it replaced, and the
// next run would see the registry recorded and leave those entries alone.
func (r *Registries) Commit(ok func(key *lockfile.Registry) bool) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	changed := false
	for _, key := range r.pending {
		if !ok(key) {
			continue
		}
		// Recording again what the lock file already says, as --force does, isn't
		// a change. A record without the commit is: the commit has to be written.
		if old := r.lf.FindRegistry(key); old != nil && old.Commit == key.Commit && old.Checksum == key.Checksum && old.ChecksumAlgorithm == key.ChecksumAlgorithm {
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
//
// A registry that failed to install isn't tried again: it fails the same way for every
// package from it. A private registry the machine can't reach would otherwise be
// requested once for each of its packages.
func (r *Registries) Get(ctx context.Context, logger *slog.Logger, name string) (*registry.Config, error) {
	rgst, ok := r.cfg.Registries[name]
	if !ok || rgst == nil {
		return nil, errRegistryNotFound
	}
	r.mu.Lock()
	inst, ok := r.installs[name]
	if !ok {
		inst = &registryInstall{}
		r.installs[name] = inst
	}
	r.mu.Unlock()
	// Installed outside the lock, so that other registries install meanwhile. The
	// installation records a checksum in pending, which takes the lock.
	inst.once.Do(func() {
		content, err := r.install(ctx, logger, rgst)
		if err != nil {
			inst.err = fmt.Errorf("install the registry: %w", slogerr.With(err, "registry_name", name))
			return
		}
		// The index of the packages by name is built on first use, without a lock.
		// Building it here, before anyone else sees the content, keeps the packages
		// resolved in parallel from building it at once.
		content.Packages(logger)
		inst.content = content
	})
	return inst.content, inst.err
}

func (r *Registries) install(ctx context.Context, logger *slog.Logger, rgst *aqua.Registry) (*registry.Config, error) {
	key := LockFileRegistry(rgst)
	if key == nil {
		return r.installer.InstallRegistry(ctx, logger, rgst, r.cfgFilePath, checksum.New()) //nolint:wrapcheck
	}
	recorded, err := r.recorded(key)
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
	if !r.record {
		return r.read(ctx, logger, rgst, recorded, chks.Get(id))
	}
	commit, err := r.commit(ctx, rgst, recorded)
	if err != nil {
		return nil, err
	}
	content, err := r.downloader.DownloadRegistry(ctx, logger, rgst, commit, chks)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	if recorded == nil {
		downloaded := chks.Get(id)
		if downloaded == nil {
			return nil, errNoRegistryChecksum
		}
		key.Commit = commit
		key.Checksum = strings.ToLower(downloaded.Checksum)
		key.ChecksumAlgorithm = downloaded.Algorithm
		r.mu.Lock()
		r.pending[rgst.Name] = key
		r.mu.Unlock()
	}
	return content, nil
}

// read reads a registry for install, exec and the update command, which record nothing.
func (r *Registries) read(ctx context.Context, logger *slog.Logger, rgst *aqua.Registry, recorded *lockfile.Registry, chk *checksum.Checksum) (*registry.Config, error) {
	if recorded == nil {
		// Only the update command gets here: recorded fails for anyone else.
		return r.installer.InstallRegistry(ctx, logger, rgst, r.cfgFilePath, r.refChecksums) //nolint:wrapcheck
	}
	// Install and exec read the registry at the recorded commit. The ref may have
	// moved on, and the file named by it is shared with every configuration.
	content, err := r.installer.InstallLockedRegistry(ctx, logger, rgst, recorded.Commit, chk)
	if err != nil {
		// The commit may have become unavailable, the repository deleted for one.
		// The hint is an attribute: slogerr keeps attributes, not the messages of
		// errors wrapped around one of its own.
		return nil, slogerr.With(err, "registry_commit", recorded.Commit, "hint", commitUnavailableHint) //nolint:wrapcheck
	}
	return content, nil
}

// commit is the commit the lock update command downloads a registry from: the recorded one,
// or, when there is none to keep, the one the ref points at now.
func (r *Registries) commit(ctx context.Context, rgst *aqua.Registry, recorded *lockfile.Registry) (string, error) {
	if recorded != nil {
		return recorded.Commit, nil
	}
	commit, _, err := r.resolver.GetCommitSHA1(ctx, rgst.RepoOwner, rgst.RepoName, rgst.Ref, "")
	if err != nil {
		return "", fmt.Errorf("resolve the registry's ref to a commit: %w", slogerr.With(err, "registry_ref", rgst.Ref))
	}
	return commit, nil
}

// recorded is the commit and the checksum the registry file is read at and checked against,
// or nil when the lock update command is to record them: there are none yet, or --force
// replaces them.
//
// A record without the commit, which a lock file written before the commit was recorded
// has, counts as none: the file can't be downloaded at what was recorded.
func (r *Registries) recorded(key *lockfile.Registry) (*lockfile.Registry, error) {
	recorded := r.lf.FindRegistry(key)
	if recorded != nil && recorded.Commit == "" {
		recorded = nil
	}
	if r.record && r.force != nil && r.force(key) {
		return nil, nil //nolint:nilnil
	}
	if recorded == nil {
		if r.record || r.readRef {
			return nil, nil //nolint:nilnil
		}
		return nil, ErrRegistryNotInLockFile
	}
	if !strongAlgorithm(recorded.ChecksumAlgorithm) {
		return nil, slogerr.With(ErrWeakRegistryChecksum, "checksum_algorithm", recorded.ChecksumAlgorithm) //nolint:wrapcheck
	}
	return recorded, nil
}

// strongAlgorithm says whether a checksum of the algorithm can stand for a registry file.
func strongAlgorithm(algorithm string) bool {
	return algorithm == "sha256" || algorithm == "sha512"
}

// Definitions is a package resolved for every environment it supports.
type Definitions struct {
	Packages []*lockfile.Package
	// PkgInfo and SupportedEnvs are what Packages were resolved from, for a package
	// whose checksums aren't in the registry: one from a registry other than the
	// standard one, or one aqua-registry-g2 serves as a template. Fetching them needs
	// these. Both are empty for a registry.json aqua-registry-g2 generated, whose
	// entries carry their checksums already.
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

type fetchFunc func(ctx context.Context, logger *slog.Logger, pkgName, version string, supportedEnvs []string) (*g2.Resolved, error)

func resolveWith(ctx context.Context, logger *slog.Logger, fetch fetchFunc, rgsts *Registries, pkg *aqua.Package) (*Definitions, error) {
	if IsStandard(rgsts.Config().Registries[pkg.Registry]) {
		if fetch == nil {
			return nil, errNoG2
		}
		supportedEnvs := supportedEnvsOf(rgsts.Config())
		resolved, err := fetch(ctx, logger, pkg.Name, pkg.Version, supportedEnvs)
		if err != nil {
			return nil, err
		}
		defs := &Definitions{Packages: resolved.Packages}
		if resolved.PkgInfo != nil {
			// A package served as a template, which was expanded here for the
			// environments the configuration supports.
			defs.PkgInfo = resolved.PkgInfo
			defs.SupportedEnvs = supportedEnvs
		}
		return defs, nil
	}
	return fromRegistry(ctx, logger, rgsts, pkg)
}

// supportedEnvsOf is the environments a configuration locks, or nil for every environment
// a package supports.
func supportedEnvsOf(cfg *aqua.Config) []string {
	if cfg == nil || cfg.Checksum == nil {
		return nil
	}
	return cfg.Checksum.SupportedEnvs
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
	supportedEnvs := supportedEnvsOf(cfg)
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
