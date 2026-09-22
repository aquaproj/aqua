package settings_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/settings"
	"github.com/google/go-cmp/cmp"
)

func TestRead(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name string
		// content is written to the settings file unless noFile is true.
		content string
		noFile  bool
		exp     *settings.Settings
		isErr   bool
	}{
		{
			name:   "no settings file",
			noFile: true,
			exp:    &settings.Settings{},
		},
		{
			name: "an empty settings file",
			exp:  &settings.Settings{},
		},
		{
			name:    "a broken settings file",
			content: "checksum: [\n",
			isErr:   true,
		},
		{
			name: "every setting",
			content: `checksum:
  enabled: true
  require: true
  enforce: true
  enforce_require: true
lazy_install: false
policy:
  enabled: false
  config: /home/foo/policy.yaml
tracking: false
progress_bar: true
log:
  level: debug
  color: always
max_parallelism: 10
global_config: /home/foo/aqua-global.yaml
root_dir: /home/foo/.aqua
keyring: true
`,
			exp: &settings.Settings{
				Checksum: &settings.Checksum{
					Enabled:        new(true),
					Require:        new(true),
					Enforce:        new(true),
					EnforceRequire: new(true),
				},
				Log: &settings.Log{
					Level: "debug",
					Color: "always",
				},
				Policy: &settings.Policy{
					Enabled: new(false),
					Config:  "/home/foo/policy.yaml",
				},
				GlobalConfig:   "/home/foo/aqua-global.yaml",
				RootDir:        "/home/foo/.aqua",
				LazyInstall:    new(false),
				Tracking:       new(false),
				ProgressBar:    new(true),
				Keyring:        new(true),
				MaxParallelism: 10,
			},
		},
		{
			// The environment variables are layered on top of these values, so
			// a setting which is written down as false must not look absent.
			name: "false isn't absent",
			content: `checksum:
  enabled: false
progress_bar: false
`,
			exp: &settings.Settings{
				Checksum: &settings.Checksum{
					Enabled: new(false),
				},
				ProgressBar: new(false),
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			p := filepath.Join(t.TempDir(), "config.yaml")
			if !d.noFile {
				if err := os.WriteFile(p, []byte(d.content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			s, err := settings.Read(p)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if diff := cmp.Diff(d.exp, s); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

// TestSettings_getters checks the value each getter falls back to when the
// setting is absent, because that fallback is what aqua does when there is no
// settings file at all.
func TestSettings_getters(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name string
		s    *settings.Settings
		exp  map[string]any
	}{
		{
			name: "absent",
			s:    &settings.Settings{},
			exp: map[string]any{
				"lazy_install":             true,
				"tracking":                 true,
				"progress_bar":             false,
				"keyring":                  false,
				"checksum.enabled":         false,
				"checksum.require":         false,
				"checksum.enforce":         false,
				"checksum.enforce_require": false,
				"log.level":                "",
				"log.color":                "",
				"policy.enabled":           true,
				"policy.config":            "",
			},
		},
		{
			name: "set",
			s: &settings.Settings{
				Checksum: &settings.Checksum{
					Enabled:        new(true),
					Require:        new(true),
					Enforce:        new(true),
					EnforceRequire: new(true),
				},
				Log: &settings.Log{
					Level: "debug",
					Color: "never",
				},
				Policy: &settings.Policy{
					Enabled: new(false),
					Config:  "/home/foo/policy.yaml",
				},
				LazyInstall: new(false),
				Tracking:    new(false),
				ProgressBar: new(true),
				Keyring:     new(true),
			},
			exp: map[string]any{
				"lazy_install":             false,
				"tracking":                 false,
				"progress_bar":             true,
				"keyring":                  true,
				"checksum.enabled":         true,
				"checksum.require":         true,
				"checksum.enforce":         true,
				"checksum.enforce_require": true,
				"log.level":                "debug",
				"log.color":                "never",
				"policy.enabled":           false,
				"policy.config":            "/home/foo/policy.yaml",
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			got := map[string]any{
				"lazy_install":             d.s.GetLazyInstall(),
				"tracking":                 d.s.GetTracking(),
				"progress_bar":             d.s.GetProgressBar(),
				"keyring":                  d.s.GetKeyring(),
				"checksum.enabled":         d.s.Checksum.GetEnabled(),
				"checksum.require":         d.s.Checksum.GetRequire(),
				"checksum.enforce":         d.s.Checksum.GetEnforce(),
				"checksum.enforce_require": d.s.Checksum.GetEnforceRequire(),
				"log.level":                d.s.Log.GetLevel(),
				"log.color":                d.s.Log.GetColor(),
				"policy.enabled":           d.s.Policy.GetEnabled(),
				"policy.config":            d.s.Policy.GetConfig(),
			}
			if diff := cmp.Diff(d.exp, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
