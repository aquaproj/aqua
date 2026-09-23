package util_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
)

// aquaEnvs are the environment variables SetParam reads.
// The test unsets all of them before each case, otherwise the environment of
// whoever runs the test decides the result.
var aquaEnvs = []string{ //nolint:gochecknoglobals
	"AQUA_CHECKSUM",
	"AQUA_DISABLE_LAZY_INSTALL",
	"AQUA_DISABLE_POLICY",
	"AQUA_DISABLE_TRACKING",
	"AQUA_ENFORCE_CHECKSUM",
	"AQUA_ENFORCE_REQUIRE_CHECKSUM",
	"AQUA_GLOBAL_CONFIG",
	"AQUA_KEYRING_ENABLED",
	"AQUA_LOG_COLOR",
	"AQUA_LOG_LEVEL",
	"AQUA_MAX_PARALLELISM",
	"AQUA_POLICY_CONFIG",
	"AQUA_PROGRESS_BAR",
	"AQUA_REQUIRE_CHECKSUM",
	"AQUA_ROOT_DIR",
	"XDG_DATA_HOME",
}

// result is the part of config.Param the settings file and the environment
// variables decide.
type result struct {
	LogLevel               string
	RootDir                string
	MaxParallelism         int
	GlobalConfigFilePaths  []string
	PolicyConfigFilePaths  []string
	ProgressBar            bool
	DisableLazyInstall     bool
	DisablePolicy          bool
	DisableTracking        bool
	Checksum               bool
	RequireChecksum        bool
	EnforceChecksum        bool
	EnforceRequireChecksum bool
}

const allSettings = `checksum:
  enabled: true
  require: true
  enforce: true
  enforce_require: true
lazy_install: false
policy:
  config: /home/foo/policy.yaml
tracking: false
progress_bar: true
log:
  level: debug
  color: never
max_parallelism: 10
global_config: /home/foo/aqua-global.yaml
root_dir: /home/foo/.aqua
`

// TestSetParam runs sequentially because it sets environment variables with
// t.Setenv, which forbids t.Parallel.
func TestSetParam(t *testing.T) { //nolint:funlen,paralleltest
	data := []struct {
		name string
		// settings is written to the settings file unless noFile is true.
		settings string
		noFile   bool
		env      map[string]string
		args     *cliargs.GlobalArgs
		exp      *result
	}{
		{
			name:   "no settings file",
			noFile: true,
			exp: &result{
				MaxParallelism:        5,
				GlobalConfigFilePaths: []string{},
				PolicyConfigFilePaths: []string{},
			},
		},
		{
			name:     "the settings file",
			settings: allSettings,
			exp: &result{
				LogLevel:               "debug",
				RootDir:                "/home/foo/.aqua",
				MaxParallelism:         10,
				GlobalConfigFilePaths:  []string{"/home/foo/aqua-global.yaml"},
				PolicyConfigFilePaths:  []string{"/home/foo/policy.yaml"},
				ProgressBar:            true,
				DisableLazyInstall:     true,
				DisableTracking:        true,
				Checksum:               true,
				RequireChecksum:        true,
				EnforceChecksum:        true,
				EnforceRequireChecksum: true,
			},
		},
		{
			name:     "the environment variables beat the settings file",
			settings: allSettings,
			env: map[string]string{
				"AQUA_CHECKSUM":                 "false",
				"AQUA_DISABLE_LAZY_INSTALL":     "false",
				"AQUA_DISABLE_TRACKING":         "false",
				"AQUA_ENFORCE_CHECKSUM":         "false",
				"AQUA_ENFORCE_REQUIRE_CHECKSUM": "false",
				"AQUA_GLOBAL_CONFIG":            "/home/bar/aqua-global.yaml",
				"AQUA_KEYRING_ENABLED":          "false",
				"AQUA_LOG_LEVEL":                "warn",
				"AQUA_MAX_PARALLELISM":          "20",
				"AQUA_POLICY_CONFIG":            "/home/bar/policy.yaml",
				"AQUA_PROGRESS_BAR":             "false",
				"AQUA_REQUIRE_CHECKSUM":         "false",
				"AQUA_ROOT_DIR":                 "/home/bar/.aqua",
			},
			// AQUA_LOG_LEVEL reaches SetParam through the flag --log-level,
			// which urfave/cli fills from the environment variable.
			args: &cliargs.GlobalArgs{LogLevel: "warn"},
			exp: &result{
				LogLevel:              "warn",
				RootDir:               "/home/bar/.aqua",
				MaxParallelism:        20,
				GlobalConfigFilePaths: []string{"/home/bar/aqua-global.yaml"},
				PolicyConfigFilePaths: []string{"/home/bar/policy.yaml"},
			},
		},
		{
			name:     "the flag beats the settings file",
			settings: "log:\n  level: debug\n",
			args:     &cliargs.GlobalArgs{LogLevel: "error"},
			exp: &result{
				LogLevel:              "error",
				MaxParallelism:        5,
				GlobalConfigFilePaths: []string{},
				PolicyConfigFilePaths: []string{},
			},
		},
		{
			// The settings file says policy.enabled, the environment variable
			// says AQUA_DISABLE_POLICY, and a disabled policy means aqua never
			// looks at the policy files.
			name:     "the policy is disabled",
			settings: "policy:\n  enabled: false\n  config: /home/foo/policy.yaml\n",
			exp: &result{
				MaxParallelism:        5,
				GlobalConfigFilePaths: []string{},
				DisablePolicy:         true,
			},
		},
	}
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		devNull.Close()
	})
	for _, d := range data { //nolint:paralleltest
		t.Run(d.name, func(t *testing.T) {
			home := setupEnv(t, d.env, d.settings, d.noFile)
			if d.exp.RootDir == "" {
				d.exp.RootDir = filepath.Join(home, ".local", "share", "aquaproj-aqua")
			}
			args := d.args
			if args == nil {
				args = &cliargs.GlobalArgs{}
			}
			param := &config.Param{}
			if err := util.SetParam(args, slogutil.New(&slogutil.InputNew{Out: devNull}), param, "v2.0.0"); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(d.exp, newResult(param)); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

// setupEnv gives the test case a home directory of its own, with the settings
// file in it unless noFile is true, and returns the home directory.
func setupEnv(t *testing.T, env map[string]string, settings string, noFile bool) string {
	t.Helper()
	home := t.TempDir()
	for _, envName := range aquaEnvs {
		t.Setenv(envName, "")
	}
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	for k, v := range env {
		t.Setenv(k, v)
	}
	if noFile {
		return home
	}
	dir := filepath.Join(home, ".config", "aquaproj-aqua")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(settings), 0o600); err != nil {
		t.Fatal(err)
	}
	return home
}

func newResult(param *config.Param) *result {
	return &result{
		LogLevel:               param.LogLevel,
		RootDir:                param.RootDir,
		MaxParallelism:         param.MaxParallelism,
		GlobalConfigFilePaths:  param.GlobalConfigFilePaths,
		PolicyConfigFilePaths:  param.PolicyConfigFilePaths,
		ProgressBar:            param.ProgressBar,
		DisableLazyInstall:     param.DisableLazyInstall,
		DisablePolicy:          param.DisablePolicy,
		DisableTracking:        param.DisableTracking,
		Checksum:               param.Checksum,
		RequireChecksum:        param.RequireChecksum,
		EnforceChecksum:        param.EnforceChecksum,
		EnforceRequireChecksum: param.EnforceRequireChecksum,
	}
}
