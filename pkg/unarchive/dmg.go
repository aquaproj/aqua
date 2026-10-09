package unarchive

import (
	"context"
	"fmt"
	"io"
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

	if err := u.attach(ctx, tempFilePath, tmpMountPoint); err != nil {
		if err := os.Remove(tmpMountPoint); err != nil {
			slogerr.WithError(logger, err).Warn("remove a temporary directory created to attach a DMG file")
		}
		return err
	}
	defer func() {
		// diskutil eject is available on macOS versions without diskutil image.
		if _, err := u.executor.ExecAndOutputWhenFailure(osexec.Command(ctx, "diskutil", "eject", tmpMountPoint)); err != nil {
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

// attach attaches a DMG file.
// hdiutil is deprecated in macOS 27, so diskutil image is used if it's available.
// diskutil image is unavailable on macOS 25 or older, so hdiutil is used as a fallback.
func (u *dmgUnarchiver) attach(ctx context.Context, dmgPath, mountPoint string) error {
	if u.supportsDiskutilImage(ctx) {
		if _, err := u.executor.ExecAndOutputWhenFailure(osexec.Command(ctx, "diskutil", "image", "attach", "--readOnly", "--nobrowse", "--mountPoint", mountPoint, dmgPath)); err != nil {
			return fmt.Errorf("diskutil image attach: %w", err)
		}
		return nil
	}
	if _, err := u.executor.ExecAndOutputWhenFailure(osexec.Command(ctx, "hdiutil", "attach", dmgPath, "-mountpoint", mountPoint)); err != nil {
		return fmt.Errorf("hdiutil attach: %w", err)
	}
	return nil
}

func (u *dmgUnarchiver) supportsDiskutilImage(ctx context.Context) bool {
	cmd := osexec.Command(ctx, "diskutil", "image")
	cmd.Stdin = nil
	// Discard the usage output of diskutil on macOS versions without diskutil image.
	cmd.Stderr = io.Discard
	_, err := u.executor.ExecAndOutputWhenFailure(cmd)
	return err == nil
}
