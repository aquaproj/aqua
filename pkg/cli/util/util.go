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
	"github.com/aquaproj/aqua/v2/pkg/log"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
	"github.com/urfave/cli/v3"
)

type Param struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	LDFlags *urfave.LDFlags
	LogE    *logrus.Entry
	Runtime *runtime.Runtime
}

func SetParam(cmd *cli.Command, logE *logrus.Entry, commandName string, param *config.Param, ldFlags *urfave.LDFlags) error { //nolint:funlen,cyclop
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	param.Args = cmd.Args().Slice()
	if logLevel := cmd.String("log-level"); logLevel != "" {
		param.LogLevel = logLevel
	}
	param.ConfigFilePath = cmd.String("config")
	param.GenerateConfigFilePath = cmd.String("generate-config")
	param.Dest = cmd.String("o")
	param.OutTestData = cmd.String("out-testdata")
	param.OnlyLink = cmd.Bool("only-link")
	param.InitConfig = cmd.Bool("init")
	if commandName == "generate-registry" {
		param.InsertFile = cmd.String("i")
	} else {
		param.Insert = cmd.Bool("i")
	}
	param.All = cmd.Bool("all")
	param.Global = cmd.Bool("g")
	param.Detail = cmd.Bool("detail")
	param.Prune = cmd.Bool("prune")
	param.CosignDisabled = cmd.Bool("disable-cosign")
	param.GitHubArtifactAttestationDisabled = cmd.Bool("disable-github-artifact-attestation")
	param.SLSADisabled = cmd.Bool("disable-slsa")
	param.Limit = int(cmd.Int("limit"))
	param.SelectVersion = cmd.Bool("select-version")
	param.Installed = cmd.Bool("installed")
	param.ShowVersion = cmd.Bool("version")
	param.File = cmd.String("f")
	if cmd := cmd.String("cmd"); cmd != "" {
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
	param.Deep = cmd.Bool("deep")
	param.Pin = cmd.Bool("pin")
	param.OnlyPackage = cmd.Bool("only-package")
	param.OnlyRegistry = cmd.Bool("only-registry")
	param.PWD = wd
	param.ProgressBar = os.Getenv("AQUA_PROGRESS_BAR") == "true"
	param.Tags = parseTags(strings.Split(cmd.String("tags"), ","))
	param.ExcludedTags = parseTags(strings.Split(cmd.String("exclude-tags"), ","))

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
