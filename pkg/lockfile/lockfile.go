// Package lockfile reads and writes aqua-lock.json.
//
// The lock file records everything needed to install a package: not just its
// checksum, the way aqua-checksums.json did, but the asset name, the format, the
// files inside the archive and how it is signed. aqua can therefore install without
// consulting a registry at all, which is what stops a private registry from blocking
// the packages that don't come from it.
//
// It is generated and updated by aqua. Editing it by hand isn't expected.
package lockfile

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

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
	Packages []*Package `json:"packages"`
}

// Metadata tells whoever opens the file what it is.
type Metadata struct {
	Description string   `json:"description,omitempty"`
	References  []string `json:"references,omitempty"`
}

// Package is everything needed to install one package on one environment.
type Package struct {
	Name     string    `json:"name"`
	Version  string    `json:"version"`
	Registry *Registry `json:"registry,omitempty"`

	OS   string `json:"os"`
	Arch string `json:"arch"`
	// Variants distinguishes entries that share an os and arch but differ
	// otherwise, such as a linux/amd64 build for musl and one for glibc.
	Variants map[string]string `json:"variants,omitempty"`

	// Checksum may be empty only for the types that build from source through
	// another tool, which verifies for itself: go_install and cargo.
	Checksum          string `json:"checksum,omitempty"`
	ChecksumAlgorithm string `json:"checksum_algorithm,omitempty"`

	Type      string `json:"type"`
	RepoOwner string `json:"repo_owner,omitempty"`
	RepoName  string `json:"repo_name,omitempty"`
	Asset     string `json:"asset,omitempty"`
	URL       string `json:"url,omitempty"`
	Format    string `json:"format,omitempty"`
	// Path is the Go module path of a go_install package. Crate and Cargo describe a
	// cargo one. Neither can be derived from a release, so they are carried here the
	// same way an asset name is.
	Path  string          `json:"path,omitempty"`
	Crate string          `json:"crate,omitempty"`
	Cargo *registry.Cargo `json:"cargo,omitempty"`

	Files []*File `json:"files,omitempty"`

	// The signing configuration is aqua's own, copied through from aqua-registry
	// rather than restated, so that verification at install time reads the same
	// shape whether the package came from a lock file or a registry.
	Cosign                     *registry.Cosign                     `json:"cosign,omitempty"`
	GitHubArtifactAttestations *registry.GitHubArtifactAttestations `json:"github_artifact_attestations,omitempty"`
	Minisign                   *registry.Minisign                   `json:"minisign,omitempty"`
}

// Registry names where the entry was resolved from.
//
// The fields are spelled out rather than combined into one string. A short form
// would need a rule to parse it, and writing the parts keeps the entry's meaning
// unchanged if another registry type appears.
type Registry struct {
	Type      string `json:"type"`
	RepoOwner string `json:"repo_owner,omitempty"`
	RepoName  string `json:"repo_name,omitempty"`
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
