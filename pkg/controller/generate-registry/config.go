package genrgst

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/asset"
	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	VersionPrefix   string
	VersionFilter   *vm.Program
	AllAssetsFilter *vm.Program
	// VersionOverrides says what to do differently for some versions, and is read
	// before the fields above: the first entry whose constraint the version meets
	// answers, and the top level answers when none does.
	VersionOverrides []*VersionOverride
	Package          string
	// Spellings names the platforms a release writes in a way the parser doesn't
	// know. Without them an asset belongs to no platform, and the package is
	// generated without it rather than with a wrong name for it.
	Spellings asset.Spellings
}

// VersionOverride is what to do differently for the versions its constraint matches.
//
// Which of a release's assets is the package can change over its history. openai/codex
// publishes about 180 assets for a dozen programs in one release, and which of them is
// codex was codex-<triple> until 0.133 and codex-package-<triple> after it. One filter
// can't say that: narrowed to the newer name the older releases have no assets at all,
// and allowing both leaves the older name matching in the newer releases, where it is
// a different archive with the binary somewhere else.
//
// A registry says this with version_overrides, so this does too.
type VersionOverride struct {
	// VersionConstraints is the expression a version meets to take this entry, in the
	// form a registry writes one. It is evaluated rather than compiled, because that
	// is what registry.PackageInfo does with its own and the two have to agree.
	VersionConstraints string
	AllAssetsFilter    *vm.Program
}

type RawConfig struct {
	VersionFilter    string                `yaml:"version_filter" json:"version_filter,omitempty"`
	VersionPrefix    string                `yaml:"version_prefix" json:"version_prefix,omitempty"`
	AllAssetsFilter  string                `yaml:"all_assets_filter" json:"all_assets_filter,omitempty"`
	VersionOverrides []*RawVersionOverride `yaml:"version_overrides" json:"version_overrides,omitempty"`
	Package          string                `yaml:"name" json:"name"`
	// Replacements says how this release writes a platform, for the ones the
	// parser doesn't recognise: luau-lang/luau calls its Linux build
	// luau-ubuntu.zip. It is the same map a registry writes, so a caller that
	// already has the package's definition can hand it over as it is.
	Replacements map[string]string `yaml:"replacements" json:"replacements,omitempty"`
}

// RawVersionOverride is a VersionOverride as it is written.
type RawVersionOverride struct {
	VersionConstraint string `yaml:"version_constraint" json:"version_constraint,omitempty"`
	AllAssetsFilter   string `yaml:"all_assets_filter" json:"all_assets_filter,omitempty"`
}

func (c *Config) FromRaw(raw *RawConfig) error {
	if raw == nil {
		return nil
	}

	c.Package = raw.Package
	c.VersionPrefix = raw.VersionPrefix
	c.Spellings = raw.Replacements

	if raw.VersionFilter != "" {
		r, err := expr.CompileVersionFilter(raw.VersionFilter)
		if err != nil {
			return fmt.Errorf("compile a version expression: %w", err)
		}
		c.VersionFilter = r
	}

	if raw.AllAssetsFilter != "" {
		a, err := expr.CompileAssetFilter(raw.AllAssetsFilter)
		if err != nil {
			return fmt.Errorf("compile an asset expression: %w", err)
		}
		c.AllAssetsFilter = a
	}

	for _, rvo := range raw.VersionOverrides {
		if rvo == nil {
			continue
		}
		vo := &VersionOverride{VersionConstraints: rvo.VersionConstraint}
		if rvo.AllAssetsFilter != "" {
			a, err := expr.CompileAssetFilter(rvo.AllAssetsFilter)
			if err != nil {
				return fmt.Errorf("compile an asset expression of a version_override: %w", err)
			}
			vo.AllAssetsFilter = a
		}
		c.VersionOverrides = append(c.VersionOverrides, vo)
	}

	return nil
}

// assetFilter is the filter to use for one version.
//
// The first entry whose constraint the version meets answers, the way a registry reads
// its own version_overrides, and the top level answers when none does. An entry
// matching the version but naming no filter answers with none, which is how a version
// says the top level's filter doesn't apply to it.
func (c *Config) assetFilter(logger *slog.Logger, tag string) *vm.Program {
	semver := strings.TrimPrefix(tag, c.VersionPrefix)
	for _, vo := range c.VersionOverrides {
		ok, err := expr.EvaluateVersionConstraints(logger, vo.VersionConstraints, tag, semver)
		if err != nil {
			// The same reading registry.PackageInfo gives it: a constraint that
			// can't be evaluated says nothing, so the entry doesn't answer.
			slogerr.WithError(logger, err).Debug("evaluate the version_constraint of a version_override",
				"version_constraint", vo.VersionConstraints, "version", tag)
			continue
		}
		if ok {
			return vo.AllAssetsFilter
		}
	}
	return c.AllAssetsFilter
}

func readConfig(path string, cfg *Config) error {
	if path == "" {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open a generate configuration file: %w", err)
	}
	defer f.Close()
	raw := &RawConfig{}
	if err := yaml.NewDecoder(f).Decode(raw); err != nil {
		return fmt.Errorf("decode a generate configuration file as YAML: %w", err)
	}
	return cfg.FromRaw(raw)
}
