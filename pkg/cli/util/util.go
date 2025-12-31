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

// SetParam configures the parameter struct with values from global args, environment variables,
// and default settings. It processes command-line arguments, sets up logging, configures
// security settings, and initializes various operational parameters for aqua commands.
func SetParam(args *cliargs.GlobalArgs, logger *slogutil.Logger, param *config.Param, version string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	if args.LogLevel != "" {
		param.LogLevel = args.LogLevel
	}
	param.ConfigFilePath = args.Config
	param.CosignDisabled = args.DisableCosign
	param.GitHubArtifactAttestationDisabled = args.DisableGitHubArtifactAttestation
	param.GitHubReleaseAttestationDisabled = args.DisableGitHubReleaseAttestation
	param.SLSADisabled = args.DisableSLSA
	param.AQUAVersion = version
	param.RootDir = config.GetRootDir(osenv.New())
	homeDir, _ := os.UserHomeDir()
	param.HomeDir = homeDir
	if err := logger.SetLevel(param.LogLevel); err != nil {
		return fmt.Errorf("set log level: %w", err)
	}
	logColor := os.Getenv("AQUA_LOG_COLOR")
	if err := logger.SetColor(logColor); err != nil {
		return fmt.Errorf("set log color: %w", err)
	}
	param.MaxParallelism = config.GetMaxParallelism(os.Getenv("AQUA_MAX_PARALLELISM"), logger.Logger)
	param.GlobalConfigFilePaths = finder.ParseGlobalConfigFilePaths(wd, os.Getenv("AQUA_GLOBAL_CONFIG"))
	param.PWD = wd
	param.ProgressBar = os.Getenv("AQUA_PROGRESS_BAR") == "true"

	for _, e := range []struct {
		envName string
		target  *bool
	}{
		{"AQUA_DISABLE_LAZY_INSTALL", &param.DisableLazyInstall},
		{"AQUA_DISABLE_POLICY", &param.DisablePolicy},
		{"AQUA_CHECKSUM", &param.Checksum},
		{"AQUA_REQUIRE_CHECKSUM", &param.RequireChecksum},
		{"AQUA_ENFORCE_CHECKSUM", &param.EnforceChecksum},
		{"AQUA_ENFORCE_REQUIRE_CHECKSUM", &param.EnforceRequireChecksum},
	} {
		if err := parseBoolEnv(e.envName, e.target); err != nil {
			return err
		}
	}
	if !param.DisablePolicy {
		param.PolicyConfigFilePaths = policy.ParseEnv(os.Getenv("AQUA_POLICY_CONFIG"))
		for i, p := range param.PolicyConfigFilePaths {
			if !filepath.IsAbs(p) {
				param.PolicyConfigFilePaths[i] = filepath.Join(param.PWD, p)
			}
		}
	}
	return nil
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
