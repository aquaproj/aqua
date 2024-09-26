package unarchive

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/exec"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const FormatDMG string = "dmg"

type dmgUnarchiver struct {
	dest     string
	executor Executor
	fs       afero.Fs
}

type Executor interface {
	Exec(cmd *exec.Cmd, param *exec.ParamRun) (int, error)
	ExecAndOutputWhenFailure(cmd *exec.Cmd) (int, error)
}

func (u *dmgUnarchiver) Unarchive(ctx context.Context, logE *logrus.Entry, src *File) error {
	if err := osfile.MkdirAll(u.fs, u.dest); err != nil {
		return fmt.Errorf("create a directory: %w", err)
	}

	tempFilePath, err := src.Body.Path()
	if err != nil {
		return fmt.Errorf("get a temporary file path: %w", err)
	}

	tmpMountPoint, err := afero.TempDir(u.fs, "", "")
	if err != nil {
		return fmt.Errorf("create a temporary file: %w", err)
	}

	if _, err := u.executor.ExecAndOutputWhenFailure(exec.Command(ctx, "hdiutil", "attach", tempFilePath, "-mountpoint", tmpMountPoint)); err != nil {
		if err := u.fs.Remove(tmpMountPoint); err != nil {
			logE.WithError(err).Warn("remove a temporary directory created to attach a DMG file")
		}
		return fmt.Errorf("hdiutil attach: %w", err)
	}
	defer func() {
		if _, err := u.executor.ExecAndOutputWhenFailure(exec.Command(ctx, "hdiutil", "detach", tmpMountPoint)); err != nil {
			logE.WithError(err).Warn("detach a DMG file")
		}
		if err := u.fs.Remove(tmpMountPoint); err != nil {
			logE.WithError(err).Warn("remove a temporary directory created to attach a DMG file")
		}
	}()

	if err := osfile.Copy(u.fs, tmpMountPoint, u.dest); err != nil {
		return fmt.Errorf("copy a directory: %w", err)
	}

	return nil
}
