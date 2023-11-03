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
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

type Controller struct {
	stdin              io.Reader
	stdout             io.Writer
	stderr             io.Writer
	which              WhichController
	packageInstaller   installpackage.Installer
	executor           Executor
	fs                 afero.Fs
	policyConfigReader PolicyReader
	policyConfigFinder policy.ConfigFinder
	enabledXSysExec    bool
	requireChecksum    bool
}

func New(param *config.Param, pkgInstaller installpackage.Installer, whichCtrl WhichController, executor Executor, osEnv osenv.OSEnv, fs afero.Fs, policyConfigReader PolicyReader, policyConfigFinder policy.ConfigFinder) *Controller {
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

type Executor interface {
	Exec(ctx context.Context, exePath string, args ...string) (int, error)
	ExecXSys(exePath string, args ...string) error
}

type PolicyReader interface {
	Read(policyFilePaths []string) ([]*policy.Config, error)
	Append(logE *logrus.Entry, aquaYAMLPath string, policies []*policy.Config, globalPolicyPaths map[string]struct{}) ([]*policy.Config, error)
}

type WhichController interface {
	Which(ctx context.Context, logE *logrus.Entry, param *config.Param, exeName string) (*which.FindResult, error)
}
