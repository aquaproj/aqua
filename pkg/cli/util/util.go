package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/log"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"github.com/urfave/cli/v2"
)

type Param struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	LDFlags *LDFlags
	LogE    *logrus.Entry
	Runtime *runtime.Runtime
}

type LDFlags struct {
	Version string
	Commit  string
	Date    string
}

func SetParam(c *cli.Context, logE *logrus.Entry, commandName string, param *config.Param, ldFlags *LDFlags) error { //nolint:funlen,cyclop
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	param.Args = c.Args().Slice()
	if logLevel := c.String("log-level"); logLevel != "" {
		param.LogLevel = logLevel
	}
	param.ConfigFilePath = c.String("config")
	param.Dest = c.String("o")
	param.OutTestData = c.String("out-testdata")
	param.OnlyLink = c.Bool("only-link")
	if commandName == "generate-registry" {
		param.InsertFile = c.String("i")
	} else {
		param.Insert = c.Bool("i")
	}
	param.All = c.Bool("all")
	param.Stdin = c.Bool("stdin")
	param.GitHub = &github.Option{
		Keyring: c.Bool("keyring"),
	}
	param.Global = c.Bool("g")
	param.Detail = c.Bool("detail")
	param.Prune = c.Bool("prune")
	param.CosignDisabled = c.Bool("disable-cosign")
	param.GitHubArtifactAttestationDisabled = c.Bool("disable-github-artifact-attestation")
	param.SLSADisabled = c.Bool("disable-slsa")
	param.Limit = c.Int("limit")
	param.SelectVersion = c.Bool("select-version")
	param.Installed = c.Bool("installed")
	param.ShowVersion = c.Bool("version")
	param.File = c.String("f")
	if cmd := c.String("cmd"); cmd != "" {
		param.Commands = strings.Split(cmd, ",")
	}
	param.LogColor = os.Getenv("AQUA_LOG_COLOR")
	param.AQUAVersion = ldFlags.Version
	param.AquaCommitHash = ldFlags.Commit
	param.RootDir = config.GetRootDir(osenv.New())
	homeDir, _ := os.UserHomeDir()
	param.HomeDir = homeDir
	log.SetLevel(param.LogLevel, logE)
	log.SetColor(param.LogColor, logE)
	param.MaxParallelism = config.GetMaxParallelism(os.Getenv("AQUA_MAX_PARALLELISM"), logE)
	param.GlobalConfigFilePaths = finder.ParseGlobalConfigFilePaths(wd, os.Getenv("AQUA_GLOBAL_CONFIG"))
	param.Deep = c.Bool("deep")
	param.Pin = c.Bool("pin")
	param.OnlyPackage = c.Bool("only-package")
	param.OnlyRegistry = c.Bool("only-registry")
	param.PWD = wd
	param.ProgressBar = os.Getenv("AQUA_PROGRESS_BAR") == "true"
	param.Tags = parseTags(strings.Split(c.String("tags"), ","))
	param.ExcludedTags = parseTags(strings.Split(c.String("exclude-tags"), ","))

	if a := os.Getenv("AQUA_DISABLE_LAZY_INSTALL"); a != "" {
		disableLazyInstall, err := strconv.ParseBool(a)
		if err != nil {
			return fmt.Errorf("parse the environment variable AQUA_DISABLE_LAZY_INSTALL as bool: %w", err)
		}
		param.DisableLazyInstall = disableLazyInstall
	}

	if a := os.Getenv("AQUA_DISABLE_POLICY"); a != "" {
		disablePolicy, err := strconv.ParseBool(a)
		if err != nil {
			return fmt.Errorf("parse the environment variable AQUA_DISABLE_POLICY as bool: %w", err)
		}
		param.DisablePolicy = disablePolicy
	}
	if !param.DisablePolicy {
		param.PolicyConfigFilePaths = policy.ParseEnv(os.Getenv("AQUA_POLICY_CONFIG"))
		for i, p := range param.PolicyConfigFilePaths {
			if !filepath.IsAbs(p) {
				param.PolicyConfigFilePaths[i] = filepath.Join(param.PWD, p)
			}
		}
	}
	if a := os.Getenv("AQUA_CHECKSUM"); a != "" {
		chksm, err := strconv.ParseBool(a)
		if err != nil {
			return fmt.Errorf("parse the environment variable AQUA_CHECKSUM as bool: %w", err)
		}
		param.Checksum = chksm
	}
	if a := os.Getenv("AQUA_REQUIRE_CHECKSUM"); a != "" {
		requireChecksum, err := strconv.ParseBool(a)
		if err != nil {
			return fmt.Errorf("parse the environment variable AQUA_REQUIRE_CHECKSUM as bool: %w", err)
		}
		param.RequireChecksum = requireChecksum
	}
	if a := os.Getenv("AQUA_ENFORCE_CHECKSUM"); a != "" {
		chksm, err := strconv.ParseBool(a)
		if err != nil {
			return fmt.Errorf("parse the environment variable AQUA_ENFORCE_CHECKSUM as bool: %w", err)
		}
		param.EnforceChecksum = chksm
	}
	if a := os.Getenv("AQUA_ENFORCE_REQUIRE_CHECKSUM"); a != "" {
		requireChecksum, err := strconv.ParseBool(a)
		if err != nil {
			return fmt.Errorf("parse the environment variable AQUA_ENFORCE_REQUIRE_CHECKSUM as bool: %w", err)
		}
		param.EnforceRequireChecksum = requireChecksum
	}
	return nil
}

func parseTags(tags []string) map[string]struct{} {
	tagsM := map[string]struct{}{}
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		tagsM[tag] = struct{}{}
	}
	return tagsM
}
