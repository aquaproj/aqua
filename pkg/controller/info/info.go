package info

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	rt "github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Controller struct {
	fs     afero.Fs
	finder ConfigFinder
	rt     *rt.Runtime
}

func New(fs afero.Fs, finder ConfigFinder, rt *rt.Runtime) *Controller {
	return &Controller{
		fs:     fs,
		finder: finder,
		rt:     rt,
	}
}

type Info struct {
	Version     string            `json:"version"`
	CommitHash  string            `json:"commit_hash"`
	OS          string            `json:"os"`
	Arch        string            `json:"arch"`
	AquaGOOS    string            `json:"aqua_goos,omitempty"`
	AquaGOARCH  string            `json:"aqua_goarch,omitempty"`
	PWD         string            `json:"pwd"`
	RootDir     string            `json:"root_dir"`
	Env         map[string]string `json:"env"`
	ConfigFiles []*Config         `json:"config_files"`
}

type Config struct {
	Path string `json:"path"`
}

func maskUser(s, username string) string {
	return strings.ReplaceAll(s, username, "(USER)")
}

func getCurrentUserName() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("get a current user: %w", err)
	}
	userName := currentUser.Username
	if runtime.GOOS == "windows" {
		// On Windows, the user name is in the form of "domain\user".
		domain, userName, ok := strings.Cut(userName, "\\")
		if ok {
			return userName, nil
		}
		return domain, nil
	}
	return userName, nil
}

func (c *Controller) Info(_ context.Context, _ *logrus.Entry, param *config.Param) error { //nolint:funlen
	userName, err := getCurrentUserName()
	if err != nil {
		return fmt.Errorf("get a current user name: %w", err)
	}

	filePaths := c.finder.Finds(param.PWD, param.ConfigFilePath)
	cfgs := make([]*Config, len(filePaths))
	for i, filePath := range filePaths {
		cfgs[i] = &Config{
			Path: maskUser(filePath, userName),
		}
	}

	info := &Info{
		Version:     param.AQUAVersion,
		CommitHash:  param.AquaCommitHash,
		PWD:         maskUser(param.PWD, userName),
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		RootDir:     maskUser(param.RootDir, userName),
		ConfigFiles: cfgs,
		Env:         map[string]string{},
	}

	if c.rt.GOOS != runtime.GOOS {
		info.AquaGOOS = c.rt.GOOS
	}
	if c.rt.GOARCH != runtime.GOARCH {
		info.AquaGOARCH = c.rt.GOARCH
	}

	envs := []string{
		"AQUA_CONFIG",
		"AQUA_DISABLE_LAZY_INSTALL",
		"AQUA_DISABLE_COSIGN",
		"AQUA_DISABLE_POLICY",
		"AQUA_DISABLE_SLSA",
		"AQUA_DISABLE_GITHUB_ARTIFACT_ATTESTATION",
		"AQUA_EXPERIMENTAL_X_SYS_EXEC",
		"AQUA_GENERATE_WITH_DETAIL",
		"AQUA_GLOBAL_CONFIG",
		"AQUA_GOARCH",
		"AQUA_GOOS",
		"AQUA_KEYRING_ENABLED",
		"AQUA_GHTKN_ENABLED",
		"AQUA_LOG_COLOR",
		"AQUA_LOG_LEVEL",
		"AQUA_MAX_PARALLELISM",
		"AQUA_POLICY_CONFIG",
		"AQUA_PROGRESS_BAR",
		"AQUA_REQUIRE_CHECKSUM",
		"AQUA_ROOT_DIR",
		"AQUA_X_SYS_EXEC",
	}
	for _, envName := range envs {
		if v, ok := os.LookupEnv(envName); ok {
			info.Env[envName] = maskUser(v, userName)
		}
	}
	if _, ok := os.LookupEnv("AQUA_GITHUB_TOKEN"); ok {
		info.Env["AQUA_GITHUB_TOKEN"] = "(masked)"
	} else if _, ok := os.LookupEnv("GITHUB_TOKEN"); ok {
		info.Env["GITHUB_TOKEN"] = "(masked)"
	}

	// check tokens for GHES
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.HasPrefix(parts[0], "AQUA_GITHUB_TOKEN_") {
			info.Env[parts[0]] = "(masked)"
		}
		if strings.HasPrefix(parts[0], "GITHUB_TOKEN_") {
			info.Env[parts[0]] = "(masked)"
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(info); err != nil {
		return fmt.Errorf("encode info as JSON and output it to stdout: %w", err)
	}
	return nil
}

type ConfigFinder interface {
	Find(wd, configFilePath string, globalConfigFilePaths ...string) (string, error)
	Finds(wd, configFilePath string) []string
}
