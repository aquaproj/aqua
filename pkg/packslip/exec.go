package packslip

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

type CommandExecutor interface {
	ExecStderr(cmd *exec.Cmd) (int, error)
}

type Executor interface {
	Verify(ctx context.Context, logger *slog.Logger, param *ParamVerify) error
}

type ExecutorImpl struct {
	executor CommandExecutor
	exePath  string
}

func NewExecutor(logger *slog.Logger, executor CommandExecutor, param *config.Param) (*ExecutorImpl, error) {
	rt := runtime.NewR(context.Background())
	pkg := Package()
	pkg.PackageInfo.OverrideByRuntime(rt)
	supported, err := pkg.PackageInfo.CheckSupported(rt, rt.Env())
	if err != nil {
		return nil, fmt.Errorf("check if the package is supported in the environment: %w", err)
	}
	if !supported {
		logger.Debug("packslip doesn't support the environment")
		return nil, nil //nolint:nilnil
	}
	exePath, err := pkg.ExePath(param.RootDir, pkg.PackageInfo.GetFiles()[0], rt)
	if err != nil {
		return nil, fmt.Errorf("get an executable file path of packslip: %w", err)
	}
	return &ExecutorImpl{
		executor: executor,
		exePath:  exePath,
	}, nil
}

var (
	// ErrVerify says packslip ran and didn't accept what it was given.
	ErrVerify = errors.New("verify with packslip")
	// errUnsupportedEnv is returned when NewExecutor returned a nil executor
	// because packslip doesn't support the host platform.
	errUnsupportedEnv = errors.New("packslip doesn't support this environment, so aqua can't verify the package with packslip")
	errNoExecutor     = errors.New("the packslip executor is nil")
)

// args are the arguments of `packslip verify`.
//
// packslip verifies offline: the bundle and the artifact are both local files, so a
// failure is a verdict on them and running it again would give the same answer.
func (p *ParamVerify) args() []string {
	args := []string{
		"verify",
		p.BundlePath,
		"--artifact",
		p.ArtifactPath,
	}
	for _, opt := range []struct {
		flag  string
		value string
	}{
		{"--pubkey", p.PublicKey},
		{"--identity", p.Identity},
		{"--identity-prefix", p.IdentityPrefix},
		{"--issuer", p.Issuer},
		{"--pin", p.Pin},
	} {
		if opt.value != "" {
			args = append(args, opt.flag, opt.value)
		}
	}
	if p.AllowUnlogged {
		args = append(args, "--allow-unlogged")
	}
	return args
}

func (e *ExecutorImpl) Verify(ctx context.Context, logger *slog.Logger, param *ParamVerify) error {
	if e == nil {
		return errUnsupportedEnv
	}
	if e.executor == nil {
		return errNoExecutor
	}
	args := param.args()
	cmd := osexec.Command(ctx, e.exePath, args...)
	cmd.Args[0] = pkgName
	if _, err := e.executor.ExecStderr(cmd); err != nil {
		if _, ok := errors.AsType[*exec.ExitError](err); !ok {
			// packslip never started, so nothing was verified either way.
			return fmt.Errorf("run packslip: %w", err)
		}
		slogerr.WithError(logger, err).Warn("execute packslip",
			"exe", e.exePath,
			"args", strings.Join(args, " "))
		return ErrVerify
	}
	return nil
}
