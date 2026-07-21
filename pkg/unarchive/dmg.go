package unarchive

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/otiai10/copy"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

const FormatDMG string = "dmg"

type dmgUnarchiver struct {
	dest     string
	executor Executor
}

type Executor interface {
	ExecAndOutputWhenFailure(cmd *osexec.Cmd) (int, error)
}

func (u *dmgUnarchiver) Unarchive(ctx context.Context, logger *slog.Logger, src *File) error {
	if err := osfile.MkdirAll(u.dest); err != nil {
		return fmt.Errorf("create a directory: %w", err)
	}

	tempFilePath, err := src.Body.Path()
	if err != nil {
		return fmt.Errorf("get a temporary file path: %w", err)
	}

	tmpMountPoint, err := os.MkdirTemp("", "")
	if err != nil {
		return fmt.Errorf("create a temporary file: %w", err)
	}

	if _, err := u.executor.ExecAndOutputWhenFailure(osexec.Command(ctx, "hdiutil", "attach", tempFilePath, "-mountpoint", tmpMountPoint)); err != nil {
		if err := os.Remove(tmpMountPoint); err != nil {
			slogerr.WithError(logger, err).Warn("remove a temporary directory created to attach a DMG file")
		}
		return fmt.Errorf("hdiutil attach: %w", err)
	}
	defer func() {
		if _, err := u.executor.ExecAndOutputWhenFailure(osexec.Command(ctx, "hdiutil", "detach", tmpMountPoint)); err != nil {
			slogerr.WithError(logger, err).Warn("detach a DMG file")
		}
		if err := os.Remove(tmpMountPoint); err != nil {
			slogerr.WithError(logger, err).Warn("remove a temporary directory created to attach a DMG file")
		}
	}()

	if err := copy.Copy(tmpMountPoint, u.dest); err != nil {
		return fmt.Errorf("copy a directory: %w", err)
	}

	return nil
}
