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

func (ctrl *Controller) Info(ctx context.Context, logE *logrus.Entry, param *config.Param, cfgFilePath string) error { //nolint:funlen
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("get a current user: %w", err)
	}
	userName := currentUser.Username

	filePaths := ctrl.finder.Finds(param.PWD, param.ConfigFilePath)
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

	if ctrl.rt.GOOS != runtime.GOOS {
		info.AquaGOOS = ctrl.rt.GOOS
	}
	if ctrl.rt.GOARCH != runtime.GOARCH {
		info.AquaGOARCH = ctrl.rt.GOARCH
	}

	envs := []string{
		"AQUA_CONFIG",
		"AQUA_DISABLE_LAZY_INSTALL",
		"AQUA_DISABLE_POLICY",
		"AQUA_EXPERIMENTAL_X_SYS_EXEC",
		"AQUA_GENERATE_WITH_DETAIL",
		"AQUA_GLOBAL_CONFIG",
		"AQUA_GOARCH",
		"AQUA_GOOS",
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
