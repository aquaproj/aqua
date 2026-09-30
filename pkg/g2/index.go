package g2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// IndexFileName is the catalogue on aqua-registry-g2's main branch.
//
// Everything else about a package lives on its own branch, which is what keeps one
// package's history out of another's. A catalogue can't: listing what a registry
// holds is a question about all of them at once, and asking it branch by branch would
// be thousands of requests to answer what one file answers.
const IndexFileName = "index.json"

// Index is what aqua-registry-g2 holds, as a list.
//
// It exists for searching. Choosing a package means reading a name and a description
// for every package there is, before knowing which one is wanted, so that has to be
// one request rather than one per package.
type Index struct {
	Packages []*IndexPackage `json:"packages"`
}

// IndexPackage is a package as it appears when searching for one.
//
// Only what a search shows or matches on: the rest is on the package's own branch and
// is read once a package has been chosen.
type IndexPackage struct {
	Name string `json:"name"`
	// UUID is what the package's branch is named after.
	//
	// A name is not an identity. A repository can be renamed, and the name it leaves
	// behind can be taken by a different repository, so a branch named after a name is
	// a branch whose subject can change under it. The branch is named after something
	// that never changes instead, and this catalogue is what turns a name -- the one a
	// package has now, or one it used to have -- into it.
	//
	// Empty while a package's branch is still named after it. Nothing resolves through
	// this yet.
	UUID        string   `json:"uuid,omitempty"`
	Description string   `json:"description,omitempty"`
	Link        string   `json:"link,omitempty"`
	Aliases     []string `json:"aliases,omitempty"`
	SearchWords []string `json:"search_words,omitempty"`
}

// NewIndexPackage reads a package's entry out of its definition.
func NewIndexPackage(pkgName string, cfg *Config) *IndexPackage {
	pkg := &IndexPackage{Name: pkgName}
	if cfg == nil || cfg.PackageInfo == nil {
		return pkg
	}
	pkg.Description = cfg.Description
	pkg.Link = cfg.Link
	pkg.SearchWords = cfg.SearchWords
	for _, alias := range cfg.Aliases {
		if alias != nil && alias.Name != "" {
			pkg.Aliases = append(pkg.Aliases, alias.Name)
		}
	}
	return pkg
}

// ReadIndex parses an index.
func ReadIndex(r io.Reader) (*Index, error) {
	index := &Index{}
	if err := json.NewDecoder(r).Decode(index); err != nil {
		return nil, fmt.Errorf("read the index as JSON: %w", err)
	}
	return index, nil
}

// Names returns the packages the index holds.
func (i *Index) Names() map[string]struct{} {
	names := make(map[string]struct{}, len(i.Packages))
	for _, pkg := range i.Packages {
		names[pkg.Name] = struct{}{}
	}
	return names
}

// Add puts packages into the index and sorts it.
//
// Sorting by name means a package added later shows up where it belongs rather than
// at the end, so the diff of an update is the lines that changed.
func (i *Index) Add(pkgs ...*IndexPackage) {
	i.Packages = append(i.Packages, pkgs...)
	slices.SortFunc(i.Packages, func(a, b *IndexPackage) int {
		return strings.Compare(a.Name, b.Name)
	})
}

// Remove takes a package out of the index.
//
// A package leaves the index when its repository is renamed: the entry under the new
// name carries the old one as an alias, and the old entry has to go or the name is both
// a package and another package's alias -- which is a state the catalogue is checked
// for, because nothing can say which of the two answers.
func (i *Index) Remove(pkgName string) {
	i.Packages = slices.DeleteFunc(i.Packages, func(pkg *IndexPackage) bool {
		return pkg != nil && pkg.Name == pkgName
	})
}

// Marshal renders the index as it is committed.
func (i *Index) Marshal() (string, error) {
	// Indented, unlike registry.json: this one is read by whoever reviews a package
	// being added, and a single line of thousands of packages can't be reviewed.
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal the index: %w", err)
	}
	return string(b) + "\n", nil
}

// CatalogueTTL is how long a cached catalogue answers.
//
// The same day the table of other names answers for, and for the same reason: what asking
// it costs is a request, and what being late costs is a package that left the registry
// going unmentioned until tomorrow. Nothing installs any differently in the meantime,
// because what installs is the lock file.
const CatalogueTTL = TableTTL

// Catalogue returns what the registry says it holds, or nil when it can't be read.
//
// It is the only thing that answers whether the registry has a package at all. A version
// answers for itself -- fetching it either finds a file or doesn't -- but a package that
// left the registry, or one that was never in it, looks from a configuration exactly like
// a package nobody has asked for yet.
//
// Nil rather than an error, the way the table of other names is: a registry that doesn't
// publish a catalogue, or a copy of one that can't be reached, is a question that goes
// unanswered rather than a run that fails.
func (c *Client) Catalogue(ctx context.Context, logger *slog.Logger, offline bool) *Index {
	if offline {
		return c.cachedCatalogue(0)
	}
	if index := c.cachedCatalogue(CatalogueTTL); index != nil {
		return index
	}
	return c.fetchCatalogue(ctx, logger)
}

// cachedCatalogue is the catalogue on disk, when it was written no longer than ttl ago. A
// ttl of zero takes it whatever its age.
func (c *Client) cachedCatalogue(ttl time.Duration) *Index {
	if c.cache == nil {
		return nil
	}
	b := c.cache.ReadWithin(c.cache.CataloguePath(c.repoOwner, c.repoName), ttl)
	if b == nil {
		return nil
	}
	index, err := ReadIndex(bytes.NewReader(b))
	if err != nil {
		// A cached file that doesn't parse is a copy gone wrong. Fetching replaces it.
		return nil
	}
	return index
}

// fetchCatalogue reads the catalogue from the registry, once per run, and caches it.
func (c *Client) fetchCatalogue(ctx context.Context, logger *slog.Logger) *Index {
	c.catalogueOnce.Do(func() {
		b, err := c.downloadFromDefaultBranch(ctx, logger, IndexFileName)
		if err != nil {
			logger.Debug("the registry publishes no catalogue", "error", err.Error())
			return
		}
		index, err := ReadIndex(bytes.NewReader(b))
		if err != nil {
			logger.Debug("the catalogue isn't readable", "error", err.Error())
			return
		}
		c.catalogue = index
		if c.cache == nil {
			return
		}
		if err := c.cache.Write(c.cache.CataloguePath(c.repoOwner, c.repoName), b); err != nil {
			// The fetch succeeded, so failing here would throw away a good answer over
			// a copy of it.
			slogerr.WithError(logger, err).Warn("cache the catalogue")
		}
	})
	return c.catalogue
}
