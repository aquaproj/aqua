// Package util provides utility functions and types for the aqua CLI package.
// It contains shared functionality for parameter handling, configuration parsing,
// and common CLI operations used across different commands.
package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/settings"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
)

// Param holds common parameters used across CLI commands.
// It contains I/O streams, build information, logging configuration,
// and runtime information needed for command execution.
type Param struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Logger  *slogutil.Logger
	Runtime *runtime.Runtime
	Version string
}

// SetParam configures the parameter struct with values from global args, the settings file,
// environment variables, and default settings. It processes command-line arguments, sets up
// logging, configures security settings, and initializes various operational parameters for
// aqua commands.
//
// A setting can come from three places, and they beat each other in this order:
// a command line flag, an environment variable, then the settings file.
func SetParam(args *cliargs.GlobalArgs, logger *slogutil.Logger, param *config.Param, version string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	osEnv := osenv.New()
	st, err := settings.Read(settings.Path(osEnv))
	if err != nil {
		return fmt.Errorf("read the settings file: %w", err)
	}

	param.CWD = wd
	param.AQUAVersion = version
	param.ConfigFilePath = args.Config
	param.CosignDisabled = args.DisableCosign
	param.GitHubArtifactAttestationDisabled = args.DisableGitHubArtifactAttestation
	param.SLSADisabled = args.DisableSLSA
	homeDir, _ := os.UserHomeDir()
	param.HomeDir = homeDir
	param.RootDir = config.GetRootDir(osEnv, st.RootDir)
	param.MaxParallelism = config.GetMaxParallelism(os.Getenv("AQUA_MAX_PARALLELISM"), st.MaxParallelism, logger.Logger)
	param.GlobalConfigFilePaths = finder.ParseGlobalConfigFilePaths(wd, getEnv("AQUA_GLOBAL_CONFIG", st.GlobalConfig))
	param.ProgressBar = getBoolEnv("AQUA_PROGRESS_BAR", st.GetProgressBar())
	param.Keyring = getBoolEnv("AQUA_KEYRING_ENABLED", st.GetKeyring())

	if err := setLog(args, logger, param, st); err != nil {
		return err
	}
	if err := setBools(param, st); err != nil {
		return err
	}
	setPolicy(param, st)
	return nil
}

// setLog resolves the log level and the log color and applies them to the logger.
// The flag --log-level also reads AQUA_LOG_LEVEL through urfave/cli, so a non empty
// args.LogLevel means one of them was given and the settings file doesn't apply.
func setLog(args *cliargs.GlobalArgs, logger *slogutil.Logger, param *config.Param, st *settings.Settings) error {
	if args.LogLevel != "" {
		param.LogLevel = args.LogLevel
	} else {
		param.LogLevel = st.Log.GetLevel()
	}
	if err := logger.SetLevel(param.LogLevel); err != nil {
		return fmt.Errorf("set log level: %w", err)
	}
	if err := logger.SetColor(getEnv("AQUA_LOG_COLOR", st.Log.GetColor())); err != nil {
		return fmt.Errorf("set log color: %w", err)
	}
	return nil
}

// setBools resolves the boolean parameters. The settings file provides the value of
// each of them, then the environment variable overwrites it: parseBoolEnv leaves the
// target untouched when the environment variable isn't set.
//
// The settings file phrases these settings positively, so the ones whose environment
// variable disables a feature are inverted here.
func setBools(param *config.Param, st *settings.Settings) error {
	param.DisableLazyInstall = !st.GetLazyInstall()
	param.DisablePolicy = !st.Policy.GetEnabled()
	param.DisableTracking = !st.GetTracking()
	param.Checksum = st.Checksum.GetEnabled()
	param.RequireChecksum = st.Checksum.GetRequire()
	param.EnforceChecksum = st.Checksum.GetEnforce()
	param.EnforceRequireChecksum = st.Checksum.GetEnforceRequire()

	for _, e := range []struct {
		envName string
		target  *bool
	}{
		{"AQUA_DISABLE_LAZY_INSTALL", &param.DisableLazyInstall},
		{"AQUA_DISABLE_POLICY", &param.DisablePolicy},
		{"AQUA_DISABLE_TRACKING", &param.DisableTracking},
		{"AQUA_CHECKSUM", &param.Checksum},
		{"AQUA_REQUIRE_CHECKSUM", &param.RequireChecksum},
		{"AQUA_ENFORCE_CHECKSUM", &param.EnforceChecksum},
		{"AQUA_ENFORCE_REQUIRE_CHECKSUM", &param.EnforceRequireChecksum},
	} {
		if err := parseBoolEnv(e.envName, e.target); err != nil {
			return err
		}
	}
	return nil
}

// setPolicy resolves the policy file paths and makes them absolute.
func setPolicy(param *config.Param, st *settings.Settings) {
	if param.DisablePolicy {
		return
	}
	param.PolicyConfigFilePaths = policy.ParseEnv(getEnv("AQUA_POLICY_CONFIG", st.Policy.GetConfig()))
	for i, p := range param.PolicyConfigFilePaths {
		if !filepath.IsAbs(p) {
			param.PolicyConfigFilePaths[i] = filepath.Join(param.CWD, p)
		}
	}
}

// getEnv returns the environment variable envName, falling back to value from the
// settings file when the environment variable is unset or empty.
func getEnv(envName, value string) string {
	if v := os.Getenv(envName); v != "" {
		return v
	}
	return value
}

// getBoolEnv is getEnv for the environment variables that have always been compared
// with the string "true" instead of being parsed with strconv.ParseBool. They keep
// that comparison, so a value aqua doesn't understand still disables the feature
// rather than failing the command.
func getBoolEnv(envName string, value bool) bool {
	if v := os.Getenv(envName); v != "" {
		return v == "true"
	}
	return value
}

func parseBoolEnv(envName string, target *bool) error {
	a := os.Getenv(envName)
	if a == "" {
		return nil
	}
	v, err := strconv.ParseBool(a)
	if err != nil {
		return fmt.Errorf("parse the environment variable %s as bool: %w", envName, err)
	}
	*target = v
	return nil
}

// ParseTags converts a slice of tag strings into a map for fast lookup.
// It trims whitespace from each tag and filters out empty strings,
// returning a map where tag names are keys with empty struct values.
func ParseTags(tags []string) map[string]struct{} {
	tagsM := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		tagsM[tag] = struct{}{}
	}
	return tagsM
}
