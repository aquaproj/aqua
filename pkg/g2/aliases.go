package g2

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
)

// AliasesFileName is the table of other names for a package, on aqua-registry-g2's
// main branch.
//
// It is a file of its own rather than part of the catalogue. Resolving one name means
// reading the table, and the catalogue is a name and a description for every package
// there is: two orders of magnitude larger, for an answer that is a few hundred lines.
const AliasesFileName = "aliases.json"

// Aliases is the other names a package is known by, as a table to look one up in.
//
// A package's aliases are usually the name its repository had before it was renamed,
// and they are how someone whose aqua.yaml still says the old name reaches the package
// at all: this registry addresses a package by name, so the name has to be resolved
// before anything can be fetched.
//
// The table is inverted from the catalogue rather than written beside it, so that the
// two can't disagree about which package an alias names.
type Aliases struct {
	// Aliases maps each other name to the package it names.
	Aliases map[string]string `json:"aliases"`
}

// NewAliases inverts the catalogue's aliases into a table.
//
// An alias naming more than one package is a mistake in the registry, which is
// reported where the catalogue is written. Here the first one wins, so that a table
// built from a catalogue with that mistake in it is still a table.
func NewAliases(index *Index) *Aliases {
	aliases := &Aliases{Aliases: map[string]string{}}
	if index == nil {
		return aliases
	}
	for _, pkg := range index.Packages {
		if pkg == nil || pkg.Name == "" {
			continue
		}
		for _, alias := range pkg.Aliases {
			if alias == "" || alias == pkg.Name {
				continue
			}
			if _, ok := aliases.Aliases[alias]; ok {
				continue
			}
			aliases.Aliases[alias] = pkg.Name
		}
	}
	return aliases
}

// ReadAliases parses a table.
func ReadAliases(r io.Reader) (*Aliases, error) {
	aliases := &Aliases{}
	if err := json.NewDecoder(r).Decode(aliases); err != nil {
		return nil, fmt.Errorf("read the aliases as JSON: %w", err)
	}
	return aliases, nil
}

// Resolve returns the package an alias names, or the name itself when it is not one.
//
// A name that is both a package and an alias of another is a mistake in the registry,
// and it isn't settled here: the table holds only aliases, so a caller that has
// already found a package under the name never asks.
func (a *Aliases) Resolve(name string) string {
	if a == nil {
		return name
	}
	if pkgName, ok := a.Aliases[name]; ok {
		return pkgName
	}
	return name
}

// Marshal renders the table as it is committed.
func (a *Aliases) Marshal() (string, error) {
	// Indented and sorted, like the catalogue: this is read by whoever reviews a
	// package being renamed, and Go writes a map's keys in order anyway.
	sorted := &Aliases{Aliases: make(map[string]string, len(a.Aliases))}
	for _, alias := range slices.Sorted(maps.Keys(a.Aliases)) {
		sorted.Aliases[alias] = a.Aliases[alias]
	}
	b, err := json.MarshalIndent(sorted, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal the aliases: %w", err)
	}
	return strings.TrimSuffix(string(b), "\n") + "\n", nil
}
