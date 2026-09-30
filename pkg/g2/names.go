package g2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// NamesFileName is the table that resolves a name, on aqua-registry-g2's main branch.
//
// A file of its own rather than part of the catalogue. Resolving a name means reading the
// whole table, and the catalogue is a name and a description for every package there is:
// an order of magnitude larger, for an answer that is a word.
const NamesFileName = "names.json"

// Names resolves a name to the branch holding the package, and an old name to the name the
// package has now.
//
// A package's branch is named after its id rather than after the package, so a name has to
// be resolved before anything can be fetched. It is one table for every name rather than a
// file per name, because a configuration asks for many packages at once: a file per name
// would be a request per package on every command that reads one, where a table is one
// request that answers for all of them and is cached afterwards.
//
// The aliases are usually the name a repository had before it was renamed, and they are how
// someone whose aqua.yaml still says the old name reaches the package at all. They are
// inverted from the catalogue rather than written beside it, so that the two can't disagree
// about which package a name belongs to.
type Names struct {
	// IDs maps a package's name to the id its branch is named after.
	IDs map[string]string `json:"ids"`
	// Aliases maps each other name a package is known by to the name it has now. It is
	// what rewrites a configuration, where IDs is what fetches a file.
	Aliases map[string]string `json:"aliases"`
}

// NewNames reads the table out of the catalogue.
//
// A name naming more than one package is a mistake in the registry, which is reported where
// the catalogue is written. Here the first one wins, so that a table built from a catalogue
// with that mistake in it is still a table.
func NewNames(index *Index) *Names {
	names := &Names{IDs: map[string]string{}, Aliases: map[string]string{}}
	if index == nil {
		return names
	}
	for _, pkg := range index.Packages {
		if pkg == nil || pkg.Name == "" {
			continue
		}
		names.add(pkg)
	}
	return names
}

// ReadNames parses a table.
func ReadNames(r io.Reader) (*Names, error) {
	names := &Names{}
	if err := json.NewDecoder(r).Decode(names); err != nil {
		return nil, fmt.Errorf("read the names as JSON: %w", err)
	}
	return names, nil
}

// ID returns the id of the branch holding the package this name belongs to, and false when
// the table doesn't know the name.
//
// An old name is answered for as well as a current one: it is the same package, and what
// the caller has is whatever its configuration says.
func (n *Names) ID(name string) (string, bool) {
	if n == nil {
		return "", false
	}
	if id, ok := n.IDs[name]; ok {
		return id, true
	}
	id, ok := n.IDs[n.Resolve(name)]
	return id, ok
}

// Resolve returns the package an old name belongs to, or the name itself when it is not one.
//
// A name that is both a package and an alias of another is a mistake in the registry, and it
// isn't settled here: a caller that found the package under the name it has never asks.
func (n *Names) Resolve(name string) string {
	if n == nil {
		return name
	}
	if pkgName, ok := n.Aliases[name]; ok {
		return pkgName
	}
	return name
}

// Marshal renders the table as it is committed.
func (n *Names) Marshal() (string, error) {
	// Indented and sorted: this is read by whoever reviews a package arriving or being
	// renamed, and Go writes a map's keys in order anyway.
	sorted := &Names{
		IDs:     make(map[string]string, len(n.IDs)),
		Aliases: make(map[string]string, len(n.Aliases)),
	}
	for _, name := range slices.Sorted(maps.Keys(n.IDs)) {
		sorted.IDs[name] = n.IDs[name]
	}
	for _, alias := range slices.Sorted(maps.Keys(n.Aliases)) {
		sorted.Aliases[alias] = n.Aliases[alias]
	}
	b, err := json.MarshalIndent(sorted, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal the names: %w", err)
	}
	return strings.TrimSuffix(string(b), "\n") + "\n", nil
}

// add is one package's names: the one it has, and the ones it used to have.
func (n *Names) add(pkg *IndexPackage) {
	if _, ok := n.IDs[pkg.Name]; !ok && pkg.ID != "" {
		n.IDs[pkg.Name] = pkg.ID
	}
	for _, alias := range pkg.Aliases {
		if alias == "" || alias == pkg.Name {
			continue
		}
		if _, ok := n.Aliases[alias]; ok {
			continue
		}
		n.Aliases[alias] = pkg.Name
	}
}

// TableTTL is how long a cached table answers a question about a name that isn't missing.
//
// A day. What it costs to be late is nothing: the name a configuration asks for still
// resolves, which is why there is anything to fix, so a rename noticed tomorrow is a line in
// a diff rather than an install that fails. What it costs to be early is a request every run.
const TableTTL = 24 * time.Hour

// NameTable returns the table, or nil when the registry has none.
//
// This is the table read forwards: given a name in a configuration, what does the registry
// call the package now. Fetching a file reads it the other way -- a name has to become an id
// before anything can be asked for -- and the difference decides how fresh a copy has to be.
// There a miss is the signal, so a stale copy costs only the fetch that replaces it. Here
// there is no miss: a copy written before the rename says the name isn't an alias, which is
// the same answer a name that was never renamed gets, and nothing would ever ask again. So a
// copy is taken while it is young and fetched once it isn't.
//
// Offline takes whatever copy there is. A run that can't reach the registry leaves the name
// alone, which is a subset of what it would have done rather than something else: the rules
// are the same rules, run against a table that may be missing a rename.
func (c *Client) NameTable(ctx context.Context, logger *slog.Logger, offline bool) *Names {
	if offline {
		return c.cachedNames()
	}
	if names := c.namesWithin(TableTTL); names != nil {
		return names
	}
	return c.fetchNames(ctx, logger)
}

// cachedNames is the table as the last run left it, or nothing.
//
// The registry is not asked for one here. Reaching a package the cached table already knows
// is nearly every read, and downloading the table to be told what it already says costs a
// request for each of them.
func (c *Client) cachedNames() *Names {
	return c.namesWithin(0)
}

// namesWithin is the cached table, when it was written no longer than ttl ago. A ttl of zero
// takes it whatever its age.
func (c *Client) namesWithin(ttl time.Duration) *Names {
	if c.cache == nil {
		return nil
	}
	b := c.cache.ReadWithin(c.cache.NamesPath(c.repoOwner, c.repoName), ttl)
	if b == nil {
		return nil
	}
	names, err := ReadNames(bytes.NewReader(b))
	if err != nil {
		// A cached file that doesn't parse is a copy gone wrong. Fetching replaces it.
		return nil
	}
	return names
}

// fetchNames reads the table from the registry, once per run, and caches it.
func (c *Client) fetchNames(ctx context.Context, logger *slog.Logger) *Names {
	c.namesOnce.Do(func() {
		b, err := c.downloadFromDefaultBranch(ctx, logger, NamesFileName)
		if err != nil {
			logger.Debug("the registry has no table of names", "error", err.Error())
			return
		}
		names, err := ReadNames(bytes.NewReader(b))
		if err != nil {
			logger.Debug("the table of names isn't readable", "error", err.Error())
			return
		}
		c.names = names
		if c.cache == nil {
			return
		}
		if err := c.cache.Write(c.cache.NamesPath(c.repoOwner, c.repoName), b); err != nil {
			// The fetch succeeded, so failing here would throw away a good answer over
			// a copy of it.
			slogerr.WithError(logger, err).Warn("cache the table of names")
		}
	})
	return c.names
}

// downloadFromDefaultBranch reads a file the registry keeps for itself rather than for one
// package: the catalogue, and the table of names beside it.
func (c *Client) downloadFromDefaultBranch(ctx context.Context, logger *slog.Logger, path string) ([]byte, error) {
	file, err := c.dl.DownloadGitHubContentFile(ctx, logger, &domain.GitHubContentFileParam{
		RepoOwner: c.repoOwner,
		RepoName:  c.repoName,
		// These are the registry as a whole rather than one package, so they sit on the
		// default branch.
		Ref:  DefaultBranch,
		Path: path,
	})
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", path, err)
	}
	defer file.Close()
	b, err := io.ReadAll(file.Reader())
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return b, nil
}
