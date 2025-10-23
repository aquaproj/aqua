package minisign

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/timer"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type CommandExecutor interface {
	ExecStderr(cmd *osexec.Cmd) (int, error)
}

type Executor interface {
	Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify, signature string) error
}

type ExecutorImpl struct {
	executor        CommandExecutor
	minisignExePath string
}

func NewExecutor(logE *logrus.Entry, executor CommandExecutor, param *config.Param) (*ExecutorImpl, error) {
	rt := runtime.NewR()
	pkg := Package()

	pkgInfo, err := pkg.PackageInfo.Override(logE, pkg.Package.Version, rt)
	if err != nil {
		return nil, fmt.Errorf("evaluate version constraints: %w", err)
	}
	supported, err := pkgInfo.CheckSupported(rt, rt.Env())
	if err != nil {
		return nil, fmt.Errorf("check if the package is supported in the environment: %w", err)
	}
	if !supported {
		logE.Debug("the package isn't supported in the environment")
		return nil, nil //nolint:nilnil
	}
	pkg.PackageInfo = pkgInfo
	exePath, err := pkg.ExePath(param.RootDir, pkgInfo.GetFiles()[0], rt)
	if err != nil {
		return nil, fmt.Errorf("get an executable file path of minisign: %w", err)
	}
	return &ExecutorImpl{
		executor:        executor,
		minisignExePath: exePath,
	}, nil
}

func wait(ctx context.Context, logE *logrus.Entry, retryCount int) error {
	randGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))       //nolint:gosec
	waitTime := time.Duration(randGenerator.Intn(1000)) * time.Millisecond //nolint:mnd
	logE.WithFields(logrus.Fields{
		"retry_count": retryCount,
		"wait_time":   waitTime,
	}).Info("Verification by minisign failed temporarily, retrying")
	if err := timer.Wait(ctx, waitTime); err != nil {
		return fmt.Errorf("wait running minisign: %w", err)
	}
	return nil
}

var errVerify = errors.New("verify with minisign")

func (e *ExecutorImpl) Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify, signature string) error {
	if e == nil {
		return errors.New("executor is nil")
	}
	// minisign -Vm myfile.txt -P RWQf6LRCGA9i53mlYecO4IzT51TGPpvWucNSCh1CBM0QTaLn73Y7GFO3
	args := []string{
		"-Vm",
		param.ArtifactPath,
		"-P",
		param.PublicKey,
		"-x",
		signature,
	}
	for i := range 5 {
		err := e.exec(ctx, args)
		if err == nil {
			return nil
		}
		logerr.WithError(logE, err).WithFields(logrus.Fields{
			"exe":  e.minisignExePath,
			"args": strings.Join(args, " "),
		}).Warn("execute minisign")
		if i == 4 { //nolint:mnd
			break
		}
		if err := wait(ctx, logE, i+1); err != nil {
			return err
		}
	}
	return errVerify
}

func (e *ExecutorImpl) exec(ctx context.Context, args []string) error {
	if e == nil {
		return errors.New("executor is nil")
	}
	if e.executor == nil {
		return errors.New("e.executor is nil")
	}
	cmd := osexec.Command(ctx, e.minisignExePath, args...)
	cmd.Args[0] = "minisign"
	_, err := e.executor.ExecStderr(cmd)
	return err //nolint:wrapcheck
}
