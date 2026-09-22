// Package settings reads aqua's settings file.
//
// The settings file holds user level preferences such as the log level or
// whether checksum verification is enforced. These used to be settable only
// through environment variables, which every shell has to set again and which
// are easy to forget.
//
// The package is deliberately separate from the package config: aqua's
// documentation calls aqua.yaml the configuration file, and the settings file
// has nothing to do with it beyond sharing the name config.yaml on disk.
package settings

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"go.yaml.in/yaml/v2"
)

// Settings is the content of the settings file.
// Every setting is optional, and the boolean settings are pointers so that a
// setting which is absent from the file can be told apart from one which is
// explicitly false. An absent setting falls back to aqua's default, while an
// explicit false must survive as false.
type Settings struct {
	Checksum       *Checksum `yaml:",omitempty" json:"checksum,omitempty"`
	Log            *Log      `yaml:",omitempty" json:"log,omitempty"`
	Policy         *Policy   `yaml:",omitempty" json:"policy,omitempty"`
	GlobalConfig   string    `yaml:"global_config,omitempty" json:"global_config,omitempty" jsonschema:"description=Global configuration file paths separated by the OS specific list separator"`
	RootDir        string    `yaml:"root_dir,omitempty" json:"root_dir,omitempty" jsonschema:"description=The directory where aqua installs packages"`
	LazyInstall    *bool     `yaml:"lazy_install,omitempty" json:"lazy_install,omitempty" jsonschema:"description=Install a package when its command is executed,default=true"`
	Tracking       *bool     `yaml:",omitempty" json:"tracking,omitempty" jsonschema:"description=Send anonymous usage data to aqua's maintainers,default=true"`
	ProgressBar    *bool     `yaml:"progress_bar,omitempty" json:"progress_bar,omitempty" jsonschema:"description=Show a progress bar while downloading packages,default=false"`
	Keyring        *bool     `yaml:",omitempty" json:"keyring,omitempty" jsonschema:"description=Get a GitHub access token from the secret store of the operating system,default=false"`
	MaxParallelism int       `yaml:"max_parallelism,omitempty" json:"max_parallelism,omitempty" jsonschema:"description=The maximum number of packages aqua installs in parallel,default=5"`
}

// Checksum holds the settings of the checksum verification.
type Checksum struct {
	Enabled        *bool `yaml:",omitempty" json:"enabled,omitempty" jsonschema:"description=Verify the checksums of packages,default=false"`
	Require        *bool `yaml:",omitempty" json:"require,omitempty" jsonschema:"description=Fail if a checksum isn't found in aqua-checksums.json,default=false"`
	Enforce        *bool `yaml:",omitempty" json:"enforce,omitempty" jsonschema:"description=Verify the checksums even if aqua.yaml disables the verification,default=false"`
	EnforceRequire *bool `yaml:"enforce_require,omitempty" json:"enforce_require,omitempty" jsonschema:"description=Require the checksums even if aqua.yaml doesn't require them,default=false"`
}

// Log holds the settings of aqua's logging.
type Log struct {
	Level string `yaml:",omitempty" json:"level,omitempty" jsonschema:"description=The log level,enum=debug,enum=info,enum=warn,enum=error"`
	Color string `yaml:",omitempty" json:"color,omitempty" jsonschema:"description=Whether logs are colored,enum=auto,enum=always,enum=never,enum=on,enum=off"`
}

// Policy holds the settings of the policy files, which decide which registries
// and packages aqua is allowed to install.
type Policy struct {
	Enabled *bool  `yaml:",omitempty" json:"enabled,omitempty" jsonschema:"description=Restrict registries and packages with policy files,default=true"`
	Config  string `yaml:",omitempty" json:"config,omitempty" jsonschema:"description=Policy file paths separated by the OS specific list separator"`
}

// Read reads the settings file at path.
//
// A missing file isn't an error because the settings file is optional; the
// returned Settings is then empty and every setting keeps its default. A file
// which can't be parsed is an error, though. Skipping it would silently drop
// settings such as checksum.enforce, and a user who wrote that down would run
// without the protection they asked for and never hear about it.
func Read(path string) (*Settings, error) {
	s := &Settings{}
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return s, nil
		}
		return nil, fmt.Errorf("open the settings file: %w", err)
	}
	defer f.Close()
	if err := yaml.NewDecoder(f).Decode(s); err != nil {
		if errors.Is(err, io.EOF) {
			// An empty file is a settings file where nothing is set,
			// not a broken one.
			return s, nil
		}
		return nil, fmt.Errorf("parse the settings file as YAML: %w", err)
	}
	return s, nil
}

// GetLazyInstall reports whether aqua installs a package when its command is
// executed. It is enabled by default.
func (s *Settings) GetLazyInstall() bool {
	if s == nil || s.LazyInstall == nil {
		return true
	}
	return *s.LazyInstall
}

// GetTracking reports whether aqua sends anonymous usage data.
// It is enabled by default.
func (s *Settings) GetTracking() bool {
	if s == nil || s.Tracking == nil {
		return true
	}
	return *s.Tracking
}

// GetProgressBar reports whether aqua shows a progress bar while downloading
// packages. It is disabled by default.
func (s *Settings) GetProgressBar() bool {
	return s != nil && s.ProgressBar != nil && *s.ProgressBar
}

// GetKeyring reports whether aqua gets a GitHub access token from the secret
// store of the operating system. It is disabled by default.
func (s *Settings) GetKeyring() bool {
	return s != nil && s.Keyring != nil && *s.Keyring
}

// GetEnabled reports whether aqua verifies the checksums of packages.
// It is disabled by default.
func (c *Checksum) GetEnabled() bool {
	return c != nil && c.Enabled != nil && *c.Enabled
}

// GetRequire reports whether aqua fails when a checksum isn't found.
// It is disabled by default.
func (c *Checksum) GetRequire() bool {
	return c != nil && c.Require != nil && *c.Require
}

// GetEnforce reports whether aqua verifies the checksums even if aqua.yaml
// disables the verification. It is disabled by default.
func (c *Checksum) GetEnforce() bool {
	return c != nil && c.Enforce != nil && *c.Enforce
}

// GetEnforceRequire reports whether aqua requires the checksums even if
// aqua.yaml doesn't require them. It is disabled by default.
func (c *Checksum) GetEnforceRequire() bool {
	return c != nil && c.EnforceRequire != nil && *c.EnforceRequire
}

// GetLevel returns the log level. An empty string means the default level.
func (l *Log) GetLevel() string {
	if l == nil {
		return ""
	}
	return l.Level
}

// GetColor returns the color mode of the logs.
// An empty string means the default mode.
func (l *Log) GetColor() string {
	if l == nil {
		return ""
	}
	return l.Color
}

// GetEnabled reports whether policy files restrict registries and packages.
// It is enabled by default.
func (p *Policy) GetEnabled() bool {
	if p == nil || p.Enabled == nil {
		return true
	}
	return *p.Enabled
}

// GetConfig returns the policy file paths separated by the OS specific list
// separator. An empty string means no policy file is configured.
func (p *Policy) GetConfig() string {
	if p == nil {
		return ""
	}
	return p.Config
}
