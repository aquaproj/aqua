package exec

import (
	"context"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

type Controller struct {
	stdin            io.Reader
	stdout           io.Writer
	stderr           io.Writer
	which            WhichController
	packageInstaller Installer
	executor         Executor
	fs               afero.Fs
	policyReader     PolicyReader
	enabledXSysExec  bool
	vacuum           Vacuum
}

type Vacuum interface {
	Update(pkgPath string, timestamp time.Time) error
}

type Installer interface {
	InstallPackage(ctx context.Context, logE *logrus.Entry, param *installpackage.ParamInstallPackage) error
}

func New(pkgInstaller Installer, whichCtrl WhichController, executor Executor, osEnv osenv.OSEnv, fs afero.Fs, policyReader PolicyReader, vacuum Vacuum) *Controller {
	return &Controller{
		stdin:            os.Stdin,
		stdout:           os.Stdout,
		stderr:           os.Stderr,
		packageInstaller: pkgInstaller,
		which:            whichCtrl,
		executor:         executor,
		enabledXSysExec:  getEnabledXSysExec(osEnv, runtime.GOOS),
		fs:               fs,
		policyReader:     policyReader,
		vacuum:           vacuum,
	}
}

type Executor interface {
	Exec(cmd *osexec.Cmd) (int, error)
	ExecXSys(exePath, name string, args ...string) error
}

type PolicyReader interface {
	Read(policyFilePaths []string) ([]*policy.Config, error)
	Append(logE *logrus.Entry, aquaYAMLPath string, policies []*policy.Config, globalPolicyPaths map[string]struct{}) ([]*policy.Config, error)
}

type WhichController interface {
	Which(ctx context.Context, logE *logrus.Entry, param *config.Param, exeName string) (*which.FindResult, error)
}
