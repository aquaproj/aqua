package g2

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
)

// VersionsFileName is the list of versions a package branch holds, at the root of that
// branch beside the definition.
//
// The list is also what the branch's tree says, and reading the tree is how it is read
// today. The file is for reading it in one object of a few kilobytes rather than an entry
// per version per directory, and for saying what the tree can't: when each release was
// published, and the digest of the file the registry serves for it.
const VersionsFileName = "versions.json"

// Versions is the content of a package branch's versions.json.
//
// What it is for is choosing a version without fetching one: a cooldown that won't take a
// release until it has stood for some days, a listing that stops at some date, the order of
// the versions of a package whose tags aren't semver. Each of those is a question about
// every version at once, and answering it from the files themselves would be a request per
// version.
type Versions struct {
	// Source is the commit of the package branch the list was generated from. It is what
	// says whether the list is still the branch's: a reader comparing it with the branch
	// knows, and the generator uses it to leave a file that is already current alone.
	//
	// When the list is stale the branch is right, since the branch is what aqua installs
	// from and this is a copy of what it holds.
	Source string `json:"source,omitempty"`
	// Versions is every version the registry holds, newest release first.
	Versions []*Version `json:"versions"`
}

// Version is one version the registry holds.
type Version struct {
	// Version is the release's tag, as upstream writes it rather than as the branch
	// escapes it for a directory name.
	Version string `json:"version"`
	// PublishedAt is when the release was published, as RFC 3339 in UTC. It is empty for
	// a version with no release to ask -- a package whose versions are tags -- which is
	// also what leaves it at the end of the list.
	PublishedAt string `json:"published_at,omitempty"`
	// Digest is the SHA-256 of the registry.json the registry serves for this version,
	// written as "sha256:" and the hex. A reader that has the file can tell whether what
	// it has is what the registry holds now.
	Digest string `json:"digest,omitempty"`
}

// ReadVersions reads a versions.json.
func ReadVersions(r io.Reader) (*Versions, error) {
	versions := &Versions{}
	if err := json.NewDecoder(r).Decode(versions); err != nil {
		return nil, fmt.Errorf("read the versions as JSON: %w", err)
	}
	return versions, nil
}

// Tags is the versions as the tags they are, in the order the list holds them.
func (v *Versions) Tags() []string {
	if v == nil {
		return nil
	}
	out := make([]string, 0, len(v.Versions))
	for _, version := range v.Versions {
		if version != nil && version.Version != "" {
			out = append(out, version.Version)
		}
	}
	return out
}

// Sort puts the newest release first, and a version with no date after every version that
// has one.
//
// The date is what a reader asks about, so it decides the order; the tag breaks a tie, so
// that two lists of the same versions are the same list whatever order they were read in.
func (v *Versions) Sort() {
	if v == nil {
		return
	}
	slices.SortFunc(v.Versions, func(a, b *Version) int {
		if a.PublishedAt != b.PublishedAt {
			// Reversed: the later date comes first. An empty date sorts last rather
			// than first, which is the one place the string order is the wrong way
			// round.
			switch {
			case a.PublishedAt == "":
				return 1
			case b.PublishedAt == "":
				return -1
			}
			return strings.Compare(b.PublishedAt, a.PublishedAt)
		}
		return strings.Compare(a.Version, b.Version)
	})
}
