package g2

import (
	"log/slog"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/expr"
)

// ConfigFileName is the file at the root of a package's branch.
//
// It sits on the package's own branch rather than on main, so that everything about
// one package is on one branch: what it is generated from and what was generated. The
// branch says which package it is, which is why the path carries no name.
//
// It is also why maintaining it needs no separate route. The branch is already where
// a generated registry.json is committed and reviewed, so a change to the definition
// travels the same way.
const ConfigFileName = "registry.yaml"

// Config is a package's definition, the input aqua-registry-g2 generates from.
//
// It is what generates registry.json rather than what installs it: aqua reads the
// generated files and never this. That is why it can carry a field aqua's own format
// has no place for.
type Config struct {
	*registry.PackageInfo `yaml:",inline"`
	// AllAssetsFilter drops assets that aren't the command, such as the shared
	// libraries and headers a release may ship beside it. A release doesn't say
	// which of its assets is the one to install, so this is written by hand. It
	// matches aqua gr's option of the same name.
	AllAssetsFilter string `yaml:"all_assets_filter,omitempty"`
	// AssetFilters says which assets are the command for some versions rather than
	// for all of them, and is read before AllAssetsFilter.
	AssetFilters []*AssetFilter `yaml:"asset_filters,omitempty"`
}

// AssetFilter is which assets are the command, for the versions its constraint
// matches.
//
// Which asset is the command can change over a package's history. openai/codex
// publishes about 180 assets for a dozen programs in one release, and which of them is
// codex was codex-<triple> until 0.133 and codex-package-<triple> after it. One filter
// can't say that: narrowed to the newer name the older releases have no assets at all,
// and allowing both leaves the older name matching in the newer releases, where it is a
// different archive with the binary somewhere else.
//
// Nothing here evaluates these. They are carried to the generator, which reads them the
// way a registry reads its version_overrides: the first whose constraint the version
// meets answers, and AllAssetsFilter answers when none does. This type exists so that
// the definition can hold them without this package depending on the generator.
type AssetFilter struct {
	VersionConstraint string `yaml:"version_constraint,omitempty"`
	AllAssetsFilter   string `yaml:"all_assets_filter,omitempty"`
}

// SetVersion returns the definition of one version.
//
// The overrides are ordered newest first and the first one that matches wins, so each
// carries the lower bound of the range it covers. A version matching none of them
// takes the first entry, which is what answers a version that isn't a semver at all.
//
// There is no version_constraint at the top level. It is the base the overrides
// inherit from and nothing else, so that a change to the newest definition doesn't
// have to be undone in every older override.
func (c *Config) SetVersion(logger *slog.Logger, v string) (*registry.PackageInfo, error) {
	if len(c.VersionOverrides) == 0 {
		return nil, errNoVersionOverride
	}
	for _, vo := range c.VersionOverrides {
		if c.match(logger, vo, v) {
			return c.OverrideVersion(vo), nil
		}
	}
	logger.Debug("no version_constraint matches; using the newest definition",
		"package_version", v)
	return c.OverrideVersion(c.VersionOverrides[0]), nil
}

// match reports whether an override covers a version.
//
// A constraint that can't be evaluated counts as not matching, the way aqua treats
// one, so a broken entry doesn't stop the ones below it from being read.
func (c *Config) match(logger *slog.Logger, vo *registry.VersionOverride, v string) bool {
	sv := v
	prefix := c.VersionPrefix
	if vo.VersionPrefix != nil {
		prefix = *vo.VersionPrefix
	}
	if prefix != "" {
		if !strings.HasPrefix(v, prefix) {
			return false
		}
		sv = strings.TrimPrefix(v, prefix)
	}
	matched, err := expr.EvaluateVersionConstraints(logger, vo.VersionConstraints, v, sv)
	if err != nil {
		logger.Debug("failed to evaluate the version_constraint",
			"version_constraint", vo.VersionConstraints, "error", err.Error())
		return false
	}
	return matched
}
