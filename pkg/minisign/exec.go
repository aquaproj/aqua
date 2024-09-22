package minisign

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/timer"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type CommandExecutor interface {
	Exec(ctx context.Context, exePath string, args ...string) (int, error)
}

type Executor interface {
	Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify, signature string) error
}

type ExecutorImpl struct {
	executor        CommandExecutor
	minisignExePath string
}

func NewExecutor(executor CommandExecutor, param *config.Param) (*ExecutorImpl, error) {
	rt := runtime.NewR()
	pkg := Package()
	pkg.PackageInfo.OverrideByRuntime(rt)
	exePath, err := pkg.ExePath(param.RootDir, pkg.PackageInfo.GetFiles()[0], rt)
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
	}).Info("Verification by minisign failed temporarily, retring")
	if err := timer.Wait(ctx, waitTime); err != nil {
		return fmt.Errorf("wait running minisign: %w", err)
	}
	return nil
}

func (e *ExecutorImpl) exec(ctx context.Context, args []string) error {
	mutex := cosign.GetMutex()
	mutex.Lock()
	defer mutex.Unlock()
	_, err := e.executor.Exec(ctx, e.minisignExePath, args...)
	return err //nolint:wrapcheck
}

var errVerify = errors.New("verify with minisign")

func (e *ExecutorImpl) Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify, signature string) error {
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
		if err := e.exec(ctx, args); err == nil {
			return nil
		} else {
			logerr.WithError(logE, err).WithFields(logrus.Fields{
				"exe":  e.minisignExePath,
				"args": strings.Join(args, " "),
			}).Warn("execute minisign")
		}
		if i == 4 { //nolint:mnd
			break
		}
		if err := wait(ctx, logE, i+1); err != nil {
			return err
		}
	}
	return errVerify
}
