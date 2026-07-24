package unarchive

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
)

const FormatPKG string = "pkg"

type pkgUnarchiver struct {
	dest     string
	executor Executor
}

func (u *pkgUnarchiver) Unarchive(ctx context.Context, _ *slog.Logger, src *File) error {
	if err := osfile.MkdirAll(filepath.Dir(u.dest)); err != nil {
		return fmt.Errorf("create a directory: %w", err)
	}

	tempFilePath, err := src.Body.Path()
	if err != nil {
		return fmt.Errorf("get a temporary file path: %w", err)
	}

	// pkgutil --expand-full fails with "File exists" unless the destination is
	// absent, but the caller hands over a directory it has already created.
	// Remove only succeeds on an empty directory, so a populated destination is
	// never destroyed here.
	if err := os.Remove(u.dest); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove the destination directory before expanding a pkg file: %w", err)
	}

	if _, err := u.executor.ExecAndOutputWhenFailure(osexec.Command(ctx, "pkgutil", "--expand-full", tempFilePath, u.dest)); err != nil {
		return fmt.Errorf("unarchive a pkg format file: %w", err)
	}

	return nil
}
