package exec

import (
	"context"
	"io"
	"os"
	"runtime"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

type Controller struct {
	stdin              io.Reader
	stdout             io.Writer
	stderr             io.Writer
	which              which.Controller
	packageInstaller   installpackage.Installer
	executor           Executor
	fs                 afero.Fs
	policyConfigReader policy.Reader
	policyConfigFinder policy.ConfigFinder
	enabledXSysExec    bool
	requireChecksum    bool
}

func New(param *config.Param, pkgInstaller installpackage.Installer, whichCtrl which.Controller, executor Executor, osEnv osenv.OSEnv, fs afero.Fs, policyConfigReader policy.Reader, policyConfigFinder policy.ConfigFinder) *Controller {
	return &Controller{
		stdin:              os.Stdin,
		stdout:             os.Stdout,
		stderr:             os.Stderr,
		packageInstaller:   pkgInstaller,
		which:              whichCtrl,
		executor:           executor,
		enabledXSysExec:    getEnabledXSysExec(osEnv, runtime.GOOS),
		fs:                 fs,
		policyConfigReader: policyConfigReader,
		policyConfigFinder: policyConfigFinder,
		requireChecksum:    param.RequireChecksum,
	}
}

func getEnabledXSysExec(osEnv osenv.OSEnv, goos string) bool {
	if goos == "windows" {
		return false
	}
	if osEnv.Getenv("AQUA_EXPERIMENTAL_X_SYS_EXEC") == "false" {
		return false
	}
	if osEnv.Getenv("AQUA_X_SYS_EXEC") == "false" {
		return false
	}
	return true
}

type Executor interface {
	Exec(ctx context.Context, exePath string, args ...string) (int, error)
	ExecXSys(exePath string, args ...string) error
}
