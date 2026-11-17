// Package lockfile reads and writes aqua-lock.json.
//
// The lock file records the checksum of every environment a package supports and the
// names of the commands it provides. It doesn't record where a package is downloaded
// from: the asset, the URL, the repository and how it is signed come from the registry
// every time, the way they did with aqua-checksums.json.
//
// That split is the point of the format. A lock file sits in the repository, so anyone
// who can change the repository can change it, and an entry saying where to download
// from would let them install whatever they like under a command name the user already
// runs. A checksum can't be turned that way: the registry decides what is downloaded,
// and a checksum that doesn't match it fails the install. The command names can't
// either: they choose among the packages aqua.yaml declares, which the policy has
// already let through. They are what lets aqua find the package a command belongs to
// without reading every registry, so a private registry doesn't block running a
// command from a public one.
//
// It is generated and updated by aqua. Editing it by hand isn't expected.
package lockfile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

// SchemaVersion is the version of the format aqua writes.
const SchemaVersion = "0.1.0"

// FileName is the name of the lock file, which sits beside aqua.yaml.
const FileName = "aqua-lock.json"

// LockFile is the content of aqua-lock.json.
type LockFile struct {
	SchemaVersion string    `json:"schema_version"`
	Metadata      *Metadata `json:"metadata,omitempty"`
	// Packages holds one entry per package, version, and environment. An entry is
	// not per package: the lock file carries every os and arch a package supports,
	// so that a repository shared between a Linux CI and a macOS laptop resolves
	// from the same file without either of them rewriting it.
	Packages []*Entry `json:"packages"`
	// Registries holds the checksum of every registry file a locked package is read
	// from. Where a package is downloaded from is the registry's answer, so the
	// registry file is what an attacker would change next; its checksum is what
	// keeps a changed file from being read. aqua-registry-g2 and local registries
	// have none: the first is read a package at a time from a repository whose
	// branches are protected, and the second sits in the repository beside the lock
	// file.
	Registries []*Registry `json:"registries,omitempty"`
}

// Metadata tells whoever opens the file what it is.
type Metadata struct {
	Description string   `json:"description,omitempty"`
	References  []string `json:"references,omitempty"`
}

// Entry is what the lock file records of one package on one environment.
//
// It holds nothing that says where the package comes from. See the package comment.
type Entry struct {
	Name     string    `json:"name"`
	Version  string    `json:"version"`
	Registry *Registry `json:"registry,omitempty"`

	OS   string `json:"os"`
	Arch string `json:"arch"`
	// Variants distinguishes entries that share an os and arch but differ
	// otherwise, such as a linux/amd64 build for musl and one for glibc.
	Variants map[string]string `json:"variants,omitempty"`

	// LinkedLibc is Package.LinkedLibc. It is kept because Find selects on it: a
	// statically linked build is what a machine gets when no entry's variants match.
	// Changing it can only pick another entry of the same package, whose definition
	// still comes from the registry.
	LinkedLibc string `json:"linked_libc,omitempty"`

	// Checksum may be empty only for the types that build from source through
	// another tool, which verifies for itself: go_install and cargo.
	Checksum          string `json:"checksum,omitempty"`
	ChecksumAlgorithm string `json:"checksum_algorithm,omitempty"`

	// Commands are the names of the executables the package provides on this
	// environment, sorted. They say which package a command belongs to, and nothing
	// about where the executable sits inside the package: that is the registry's.
	Commands []string `json:"commands,omitempty"`
}

// Package is everything needed to install one package on one environment.
//
// It is what a registry resolves to, not what the lock file holds: Entry is the part of
// it the lock file records. aqua-registry-g2's registry.json is a list of these.
type Package struct {
	Name     string    `json:"name"`
	Version  string    `json:"version"`
	Registry *Registry `json:"registry,omitempty"`

	OS   string `json:"os"`
	Arch string `json:"arch"`
	// Variants distinguishes entries that share an os and arch but differ
	// otherwise, such as a linux/amd64 build for musl and one for glibc.
	Variants map[string]string `json:"variants,omitempty"`

	// LinkedLibc is the libc the executables in this asset are linked against:
	// "musl", "glibc", or "static" when they need none. It is read from the binary
	// rather than from the asset's name, and is absent when the asset holds nothing
	// that can be read that way, such as a script.
	//
	// Nothing selects on it today; Variants does that. It is recorded because which
	// build a machine should prefer is a choice worth leaving open: a statically
	// linked musl build runs anywhere, while a glibc one depends on the version of
	// glibc it was built against.
	LinkedLibc string `json:"linked_libc,omitempty"`

	// Rosetta2 and WindowsARMEmulation say the entry runs the amd64 build on an arm64
	// machine. See g2.Asset.
	Rosetta2            bool `json:"rosetta2,omitempty"`
	WindowsARMEmulation bool `json:"windows_arm_emulation,omitempty"`

	// Checksum may be empty only for the types that build from source through
	// another tool, which verifies for itself: go_install and cargo.
	Checksum          string `json:"checksum,omitempty"`
	ChecksumAlgorithm string `json:"checksum_algorithm,omitempty"`

	Type      string `json:"type"`
	RepoOwner string `json:"repo_owner,omitempty"`
	RepoName  string `json:"repo_name,omitempty"`
	// RepoID is the repository's numeric id on GitHub, carried from the registry so that
	// the lock records which repository the name meant when it was locked.
	RepoID int64 `json:"repo_id,omitempty"`
	// Host is the forge instance the release is on, for a forgejo_release or
	// gitea_release entry. The other types need no such field: their host is
	// github.com, or is part of URL.
	Host   string `json:"host,omitempty"`
	Asset  string `json:"asset,omitempty"`
	URL    string `json:"url,omitempty"`
	Format string `json:"format,omitempty"`
	// Path is the Go module path of a go_install package. Crate and Cargo describe a
	// cargo one. Neither can be derived from a release, so they are carried here the
	// same way an asset name is.
	Path  string          `json:"path,omitempty"`
	Crate string          `json:"crate,omitempty"`
	Cargo *registry.Cargo `json:"cargo,omitempty"`

	Files []*File `json:"files,omitempty"`

	// Private says the repository needs a token to download from, which changes
	// where the download is asked for rather than whether it works: aqua tries the
	// anonymous URL first for a public repository, to keep a token's rate limit for
	// things that need it, and goes straight to the API for a private one.
	Private bool `json:"private,omitempty"`

	// The signing configuration is aqua's own, copied through from aqua-registry
	// rather than restated, so that verification at install time reads the same
	// shape whether the package came from a lock file or a registry. The names of
	// the signature files are rendered for this environment like everything else
	// here: an entry that kept the templates would need the replacements kept with
	// it to expand them again.
	Cosign                     *registry.Cosign                     `json:"cosign,omitempty"`
	GitHubArtifactAttestations *registry.GitHubArtifactAttestations `json:"github_artifact_attestations,omitempty"`
	Minisign                   *registry.Minisign                   `json:"minisign,omitempty"`
	Packslip                   *registry.Packslip                   `json:"packslip,omitempty"`
	SLSAProvenance             *registry.SLSAProvenance             `json:"slsa_provenance,omitempty"`
}

// Registry names where the entry was resolved from. In LockFile.Registries it also
// names one file of the registry at one ref, with that file's checksum.
//
// The fields are spelled out rather than combined into one string. A short form
// would need a rule to parse it, and writing the parts keeps the entry's meaning
// unchanged if another registry type appears.
type Registry struct {
	Type              string `json:"type"`
	RepoOwner         string `json:"repo_owner,omitempty"`
	RepoName          string `json:"repo_name,omitempty"`
	Ref               string `json:"ref,omitempty"`
	Path              string `json:"path,omitempty"`
	Checksum          string `json:"checksum,omitempty"`
	ChecksumAlgorithm string `json:"checksum_algorithm,omitempty"`
}

// SameFile says whether two registries name the same file at the same ref.
func (r *Registry) SameFile(other *Registry) bool {
	return r.Type == other.Type && r.RepoOwner == other.RepoOwner && r.RepoName == other.RepoName &&
		r.Ref == other.Ref && r.Path == other.Path
}

// FindRegistry returns the recorded checksum of the registry file rgst names, or nil.
func (lf *LockFile) FindRegistry(rgst *Registry) *Registry {
	if lf == nil {
		return nil
	}
	for _, r := range lf.Registries {
		if r != nil && r.SameFile(rgst) {
			return r
		}
	}
	return nil
}

// SetRegistry records the checksum of a registry file, replacing what was recorded for
// the same file.
func (lf *LockFile) SetRegistry(rgst *Registry) {
	lf.Registries = slices.DeleteFunc(lf.Registries, func(r *Registry) bool {
		return r == nil || r.SameFile(rgst)
	})
	lf.Registries = append(lf.Registries, rgst)
}

// RetainRegistries drops the registry files a predicate doesn't declare, and reports how
// many went.
func (lf *LockFile) RetainRegistries(declared func(rgst *Registry) bool) int {
	before := len(lf.Registries)
	lf.Registries = slices.DeleteFunc(lf.Registries, func(r *Registry) bool {
		return r == nil || !declared(r)
	})
	return before - len(lf.Registries)
}

// Entry returns what the lock file records of the package.
func (p *Package) Entry() *Entry {
	return &Entry{
		Name:              p.Name,
		Version:           p.Version,
		Registry:          p.Registry,
		OS:                p.OS,
		Arch:              p.Arch,
		Variants:          p.Variants,
		LinkedLibc:        p.LinkedLibc,
		Checksum:          p.Checksum,
		ChecksumAlgorithm: p.ChecksumAlgorithm,
		Commands:          p.commands(),
	}
}

// commands are the names of the package's files, sorted and without repeats.
//
// A definition without files has one command after all, named after the package, and
// that is the command the package is installed with. It is recorded the same way, or the
// entry would have no commands for any run of the lock update command to fill in.
func (p *Package) commands() []string {
	files := p.PackageInfo().GetFiles()
	if len(files) == 0 {
		return nil
	}
	names := make([]string, 0, len(files))
	for _, f := range files {
		names = append(names, f.Name)
	}
	slices.Sort(names)
	return slices.Compact(names)
}

// CheckCommands says whether the entry's commands can be used as they are.
//
// A command name becomes the name of a link in $(aqua root-dir)/bin, so it has to be a
// name and nothing else. The lock file can be changed by whoever can change the
// repository, and a name such as "../../.bashrc" would replace a file outside that
// directory with a link. An entry with no commands at all was written before the lock
// file recorded them, and guessing them from the package name would create the wrong
// links.
func (e *Entry) CheckCommands() error {
	if len(e.Commands) == 0 {
		return ErrNoCommands
	}
	for _, cmd := range e.Commands {
		if !validCommand(cmd) {
			return fmt.Errorf("%w: %q", ErrInvalidCommand, cmd)
		}
	}
	return nil
}

func validCommand(cmd string) bool {
	return cmd != "" && cmd != "." && cmd != ".." && !strings.ContainsAny(cmd, `/\:`) && !strings.ContainsRune(cmd, 0)
}

var (
	// ErrNoCommands is an entry written before the lock file recorded commands.
	ErrNoCommands = errors.New(`the lock file entry has no commands. Please run "aqua lock update --force"`)
	// ErrInvalidCommand is a command name that isn't a plain name.
	ErrInvalidCommand = errors.New("the lock file entry has a command name that isn't a plain name")
)

// Describes says whether the entry is the one for p's environment: the same os, arch
// and variants. The name and version aren't compared, since p is looked up by them.
func (e *Entry) Describes(p *Package) bool {
	return e.OS == p.OS && e.Arch == p.Arch && maps.Equal(e.Variants, p.Variants)
}

// File is an executable inside the package.
type File struct {
	Name string `json:"name"`
	Src  string `json:"src,omitempty"`
}

// New creates an empty lock file.
func New() *LockFile {
	return &LockFile{
		SchemaVersion: SchemaVersion,
		Metadata: &Metadata{
			Description: "This file is managed by aqua. Don't edit this file directly.",
		},
	}
}

// Read parses a lock file.
func Read(r io.Reader) (*LockFile, error) {
	lf := &LockFile{}
	if err := json.NewDecoder(r).Decode(lf); err != nil {
		return nil, fmt.Errorf("read the lock file as JSON: %w", err)
	}
	// A null entry describes nothing, and every lookup would have to step around it.
	lf.Packages = slices.DeleteFunc(lf.Packages, func(pkg *Entry) bool { return pkg == nil })
	lf.Registries = slices.DeleteFunc(lf.Registries, func(rgst *Registry) bool { return rgst == nil })
	// The commands are compared with the registry's as a sorted list. A file edited by
	// hand or merged can hold the same commands in another order.
	for _, pkg := range lf.Packages {
		slices.Sort(pkg.Commands)
		pkg.Commands = slices.Compact(pkg.Commands)
	}
	return lf, nil
}

// ReadFile reads the lock file at path, returning nil when there is none.
//
// Absent and empty are different answers. A repository with no lock file hasn't
// adopted one, and aqua still resolves its packages through registries; a lock file
// that exists is the statement that it is now the authority, and a package missing
// from it is a gap to report rather than something to look up elsewhere. Callers
// therefore have to tell the two apart, so this doesn't paper over the difference by
// returning an empty file.
func ReadFile(path string) (*LockFile, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil //nolint:nilnil
		}
		return nil, fmt.Errorf("open the lock file: %w", err)
	}
	defer f.Close()
	return Read(f)
}

// filePerm is the mode the lock file is created with. It is committed to git, so it
// is readable.
const filePerm = 0o644

// Write writes the lock file to path, sorted.
func Write(path string, lf *LockFile) error {
	lf.Sort()
	b, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal the lock file: %w", err)
	}
	if err := os.WriteFile(path, append(b, '\n'), filePerm); err != nil {
		return fmt.Errorf("write the lock file: %w", err)
	}
	return nil
}
