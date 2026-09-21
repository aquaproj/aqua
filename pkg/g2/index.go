package g2

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
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
	Name        string   `json:"name"`
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
