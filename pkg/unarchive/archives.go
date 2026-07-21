package unarchive

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/mholt/archives"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

var errEscapeDest = errors.New("the file path escapes the extraction directory")

type handler struct {
	fs       afero.Fs
	executor Executor
	dest     string
	filename string
	logger   *slog.Logger
	// sparse records that the archive contains a GNU sparse entry that Go's
	// archive/tar cannot extract. It is set during HandleFile so that unarchive
	// can fall back to the system tar command after the walk is aborted.
	sparse bool
}

func (h *handler) HandleFile(_ context.Context, f archives.FileInfo) error {
	// GNU sparse entries (PAX GNU.sparse.* or the old GNU sparse type) are not
	// reliably extractable by Go's archive/tar. Detect them before reading the
	// body -- reading a sparse body would materialize its full logical size (up
	// to many GiB of zeros) before failing -- and abort the walk so unarchive
	// can re-extract the whole archive with the system tar command.
	if isGNUSparse(f) {
		h.sparse = true
		return fs.SkipAll
	}

	dstPath := filepath.Join(h.dest, h.normalizePath(f.NameInArchive))
	if !h.withinDest(dstPath) {
		return fmt.Errorf("%w: %s", errEscapeDest, f.NameInArchive)
	}
	parentDir := filepath.Dir(dstPath)
	// Reject an entry whose parent directory resolves outside dest through a
	// symlink planted by an earlier entry. Creating the parent directories or
	// writing the entry would otherwise follow that symlink and escape dest.
	// The root entry ("./") is exempt: its dstPath equals dest, so its parent is
	// legitimately outside dest. dest itself is validated in the handlers below.
	if dstPath != filepath.Clean(h.dest) && h.escapesDest(parentDir) {
		return fmt.Errorf("%w: %s", errEscapeDest, f.NameInArchive)
	}
	if err := osfile.MkdirAll(parentDir); err != nil {
		slogerr.WithError(h.logger, err).Warn("create a directory")
		return nil
	}

	if f.IsDir() {
		return h.handleDir(dstPath, f)
	}

	if f.LinkTarget != "" {
		if f.Mode()&os.ModeSymlink != 0 {
			h.handleSymlink(dstPath, f.LinkTarget)
		}
		return nil
	}

	return h.handleRegularFile(dstPath, f)
}

func (h *handler) Unarchive(ctx context.Context, _ *slog.Logger, src *File) error {
	tempFilePath, err := src.Body.Path()
	if err != nil {
		return fmt.Errorf("get a temporary file path: %w", err)
	}
	if err := h.unarchive(ctx, src.Filename, tempFilePath); err != nil {
		return slogerr.With(err, "archived_file", tempFilePath, "archived_filename", src.Filename) //nolint:wrapcheck
	}
	return nil
}

func (h *handler) handleDir(dstPath string, f archives.FileInfo) error {
	// Guard against a directory entry whose path was planted as an escaping
	// symlink by an earlier entry; MkdirAll would otherwise follow it.
	if h.escapesDest(dstPath) {
		return fmt.Errorf("%w: %s", errEscapeDest, f.NameInArchive)
	}
	if err := h.fs.MkdirAll(dstPath, f.Mode()|0o700); err != nil { //nolint:mnd
		slogerr.WithError(h.logger, err).Warn("create a directory")
	}
	return nil
}

func (h *handler) handleRegularFile(dstPath string, f archives.FileInfo) error {
	// Refuse to write through an escaping symlink planted at dstPath itself. An
	// archive can create "pwn -> /outside" and then a regular file entry "pwn";
	// opening it with O_CREATE would follow the symlink and write outside dest.
	if h.escapesDest(dstPath) {
		return fmt.Errorf("%w: %s", errEscapeDest, f.NameInArchive)
	}

	reader, err := f.Open()
	if err != nil {
		slogerr.WithError(h.logger, err).Warn("open a file")
		return nil
	}
	defer reader.Close()

	dstFile, err := h.fs.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, f.Mode())
	if err != nil {
		slogerr.WithError(h.logger, err).Warn("create a file")
		return nil
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, reader); err != nil {
		slogerr.WithError(h.logger, err).Warn("copy a file")
	}
	return nil
}

// handleSymlink creates a symlink at dstPath pointing to target. The symlink is
// created even when target resolves outside h.dest: the symlink inode itself
// writes nothing outside the extraction directory, and such symlinks are common
// in legitimate archives (e.g. a root filesystem's "var/run -> /run"). Escaping
// through the symlink is instead prevented at write time by escapesDest, which
// rejects any later entry that would follow it out of h.dest. The caller has
// already verified dstPath's parent directory stays within h.dest.
func (h *handler) handleSymlink(dstPath, target string) {
	if err := os.Symlink(target, dstPath); err != nil {
		slogerr.WithError(h.logger, err).Warn("create a symlink", "link_target", target, "link_dest", dstPath)
	}
}

// maxSymlinkHops bounds how many symlinks escapesDest follows before giving up,
// guarding against symlink cycles. It matches the conventional MAXSYMLINKS.
const maxSymlinkHops = 255

// escapesDest reports whether writing to path would land outside h.dest once the
// symlinks along path are followed. Unlike filepath.EvalSymlinks it does not
// require path's final target to exist, so it also catches a dangling symlink
// planted at path that a later O_CREATE write would follow out of h.dest. It
// detects a symlink planted at path itself as well as an escaping symlink in any
// parent directory. h.dest is resolved too because it may itself contain
// symlinks (e.g. macOS /var -> /private/var), which would otherwise make every
// entry look like an escape.
func (h *handler) escapesDest(path string) bool {
	dest, err := filepath.EvalSymlinks(h.dest)
	if err != nil {
		dest = filepath.Clean(h.dest)
	}
	resolved, ok := h.resolveSymlinks(path, 0)
	if !ok {
		// A symlink cycle or an unreadable link: treat as unsafe.
		return true
	}
	resolved = filepath.Clean(resolved)
	if resolved == dest {
		return false
	}
	return !strings.HasPrefix(resolved, dest+string(filepath.Separator))
}

// resolveSymlinks resolves the symlinks in path, following even a dangling
// symlink whose final target does not exist yet (which filepath.EvalSymlinks
// refuses to do). Components that do not exist and are not symlinks are kept
// literally. It returns false if a symlink cycle or an unreadable link is hit.
func (h *handler) resolveSymlinks(path string, hops int) (string, bool) {
	if hops > maxSymlinkHops {
		return "", false
	}
	// Fast path: a fully existing path resolves cleanly.
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved, true
	}
	// path did not fully resolve. If path itself is a symlink (possibly
	// dangling), follow its target manually.
	if fi, err := os.Lstat(path); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return "", false
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}
		return h.resolveSymlinks(target, hops+1)
	}
	// path does not exist and is not a symlink. Resolve its parent so an escaping
	// symlink in an ancestor is still followed, then re-attach the base name.
	parent := filepath.Dir(path)
	if parent == path {
		return path, true
	}
	resolvedParent, ok := h.resolveSymlinks(parent, hops)
	if !ok {
		return "", false
	}
	return filepath.Join(resolvedParent, filepath.Base(path)), true
}

// withinDest reports whether the cleaned path is h.dest itself or located inside
// it. It is used to reject archive entries whose destination escapes the
// extraction directory (e.g. via ".." in the archived path).
func (h *handler) withinDest(p string) bool {
	p = filepath.Clean(p)
	dest := filepath.Clean(h.dest)
	if p == dest {
		return true
	}
	return strings.HasPrefix(p, dest+string(filepath.Separator))
}

func (h *handler) normalizePath(nameInArchive string) string {
	slashCount := strings.Count(nameInArchive, "/")
	backSlashCount := strings.Count(nameInArchive, "\\")
	if backSlashCount > slashCount && filepath.Separator != '\\' {
		return strings.ReplaceAll(nameInArchive, "\\", string(filepath.Separator))
	}
	return nameInArchive
}

func (h *handler) unarchive(ctx context.Context, fileName, file string) error {
	archiveFile, err := h.fs.Open(file)
	if err != nil {
		return fmt.Errorf("open a files: %w", err)
	}
	defer archiveFile.Close()

	format, input, err := archives.Identify(ctx, fileName, archiveFile)
	if err != nil {
		return fmt.Errorf("identify the format: %w", err)
	}

	if extractor, ok := format.(archives.Extractor); ok {
		if err := osfile.MkdirAll(h.dest); err != nil {
			return fmt.Errorf("create a destination directory: %w", err)
		}

		if err := extractor.Extract(ctx, input, h.HandleFile); err != nil {
			return fmt.Errorf("extract files: %w", err)
		}
		if h.sparse {
			return h.extractWithSystemTar(ctx, file)
		}
		return nil
	}
	if decomp, ok := format.(archives.Decompressor); ok {
		return h.decompress(input, decomp)
	}
	return errUnsupportedFileFormat
}

// isGNUSparse reports whether f is a GNU sparse tar entry. Such entries are
// stored either in the old GNU sparse format (Typeflag TypeGNUSparse) or the
// PAX-based GNU sparse 0.x/1.0 format (a GNU.sparse.* extended header record).
// Go's archive/tar cannot reliably extract these, so aqua falls back to the
// system tar command when one is found.
func isGNUSparse(f archives.FileInfo) bool {
	th, ok := f.Header.(*tar.Header)
	if !ok {
		return false
	}
	if th.Typeflag == tar.TypeGNUSparse {
		return true
	}
	for k := range th.PAXRecords {
		if strings.HasPrefix(k, "GNU.sparse.") {
			return true
		}
	}
	return false
}

// extractWithSystemTar extracts the archive with the system tar command. It is
// used as a fallback for GNU sparse archives, which Go's archive/tar cannot
// extract but GNU tar and bsdtar can. The system tar also restores the files as
// sparse on disk instead of materializing the holes as zeros. A tar binary on
// PATH is required; on macOS and Linux it is available by default.
//
// The destination directory has already been created by unarchive. The system
// tar auto-detects the compression (gzip, xz, zstd, ...) from the archive
// content, so no compression flag is needed.
func (h *handler) extractWithSystemTar(ctx context.Context, file string) error {
	h.logger.Warn("the archive contains GNU sparse files unsupported by the Go tar reader; falling back to the system tar command")
	if h.executor == nil {
		return errors.New("cannot extract a GNU sparse archive: no command executor is available")
	}
	cmd := osexec.Command(ctx, "tar", "-x", "-f", file, "-C", h.dest)
	if _, err := h.executor.ExecAndOutputWhenFailure(cmd); err != nil {
		return fmt.Errorf("extract a GNU sparse archive with the system tar command (a `tar` binary on PATH is required for GNU sparse archives): %w", err)
	}
	return nil
}

func (h *handler) decompress(input io.Reader, decomp archives.Decompressor) error {
	rc, err := decomp.OpenReader(input)
	if err != nil {
		return fmt.Errorf("open a decompressed file: %w", err)
	}
	defer rc.Close()
	if err := osfile.MkdirAll(h.dest); err != nil {
		return fmt.Errorf("create a directory (%s): %w", h.dest, err)
	}
	dst, err := h.fs.Create(filepath.Join(h.dest, strings.TrimSuffix(h.filename, filepath.Ext(h.filename))))
	if err != nil {
		return fmt.Errorf("create a destination file: %w", err)
	}
	defer dst.Close()
	if _, err := io.Copy(dst, rc); err != nil {
		return fmt.Errorf("copy decompressed data: %w", err)
	}
	return nil
}
