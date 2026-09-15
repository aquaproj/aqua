package unarchive

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/mholt/archives"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

var (
	errEscapeDest = errors.New("the file path escapes the extraction directory")
	// errSparseMemberName is returned when a GNU sparse entry's name cannot be
	// handed to the system tar command as a member operand without the risk of
	// selecting other entries. See validateSparseMember.
	errSparseMemberName = errors.New("the GNU sparse file name cannot be passed to the system tar command safely")
)

type handler struct {
	executor Executor
	dest     string
	filename string
	logger   *slog.Logger
	// sparseMembers records the GNU sparse entries found during the walk. Go's
	// archive/tar cannot restore them as sparse files, so after the walk
	// unarchive re-extracts only these specific members with the system tar
	// command. Recording individual members -- rather than re-extracting the
	// whole archive -- keeps every symlink, directory, and regular file under
	// the hardened Go walk's containment checks; the system tar only ever
	// extracts the members validated by extractSparseMembers, and only into a
	// staging directory outside the walk's output.
	sparseMembers []sparseMember
	// entryNames records the name of every entry seen during the walk, in
	// archive order. extractSparseMembers needs them because the system tar
	// selects members by pattern, not by literal name: a sparse member may only
	// be handed to it once no other entry can be selected by the same operand.
	entryNames []string
}

// sparseMember identifies a GNU sparse entry to hand to the system tar command.
type sparseMember struct {
	// name is the entry's name as stored in the archive. It is passed to the
	// system tar as the member to extract, and it is where the system tar writes
	// the member inside the staging directory.
	name string
	// dstPath is the normalized destination path, validated for containment
	// before the member is moved there.
	dstPath string
}

func (h *handler) HandleFile(_ context.Context, f archives.FileInfo) error {
	h.entryNames = append(h.entryNames, f.NameInArchive)
	dstPath := filepath.Join(h.dest, h.normalizePath(f.NameInArchive))
	if !h.withinDest(dstPath) {
		return fmt.Errorf("%w: %s", errEscapeDest, f.NameInArchive)
	}

	// GNU sparse entries (PAX GNU.sparse.* or the old GNU sparse type) cannot be
	// restored as sparse files by Go's archive/tar, so they are re-extracted by
	// the system tar command after the walk (see extractSparseMembers). Do not
	// read the body here -- reading a sparse body would materialize its full
	// logical size (up to many GiB of zeros). The walk is not aborted: it
	// continues so every remaining entry stays under the containment checks
	// below, and the sparse member's containment is re-validated once all
	// symlinks are planted, just before the member is moved into place.
	if isGNUSparse(f) {
		h.sparseMembers = append(h.sparseMembers, sparseMember{name: f.NameInArchive, dstPath: dstPath})
		return nil
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
	if err := os.MkdirAll(dstPath, f.Mode()|0o700); err != nil { //nolint:mnd
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

	dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, f.Mode())
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
func (h *handler) escapesDest(p string) bool {
	dest, err := filepath.EvalSymlinks(h.dest)
	if err != nil {
		dest = filepath.Clean(h.dest)
	}
	resolved, ok := h.resolveSymlinks(p, 0)
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
func (h *handler) resolveSymlinks(p string, hops int) (string, bool) {
	if hops > maxSymlinkHops {
		return "", false
	}
	// Fast path: a fully existing path resolves cleanly.
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved, true
	}
	// p did not fully resolve. If p itself is a symlink (possibly dangling),
	// follow its target manually.
	if fi, err := os.Lstat(p); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(p)
		if err != nil {
			return "", false
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(p), target)
		}
		return h.resolveSymlinks(target, hops+1)
	}
	// p does not exist and is not a symlink. Resolve its parent so an escaping
	// symlink in an ancestor is still followed, then re-attach the base name.
	parent := filepath.Dir(p)
	if parent == p {
		return p, true
	}
	resolvedParent, ok := h.resolveSymlinks(parent, hops)
	if !ok {
		return "", false
	}
	return filepath.Join(resolvedParent, filepath.Base(p)), true
}

// withinDest reports whether the cleaned path is h.dest itself or located inside
// it. It is used to reject archive entries whose destination escapes the
// extraction directory (e.g. via ".." in the archived path).
func (h *handler) withinDest(p string) bool {
	return within(h.dest, p)
}

// within reports whether the cleaned path p is dir itself or located inside it.
// It compares the paths lexically; it does not resolve symlinks (escapesDest
// does that).
func within(dir, p string) bool {
	p = filepath.Clean(p)
	dir = filepath.Clean(dir)
	if p == dir {
		return true
	}
	return strings.HasPrefix(p, dir+string(filepath.Separator))
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
	archiveFile, err := os.Open(file)
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
		if len(h.sparseMembers) > 0 {
			return h.extractSparseMembers(ctx, file)
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

// extractSparseMembers restores the archive's GNU sparse members as sparse files
// with the system tar command, which Go's archive/tar cannot do (it rejects some
// GNU sparse archives outright, e.g. "sparse file contains unreferenced data").
// Only the sparse members recorded during the walk are extracted; every other
// entry (symlinks, directories, regular files) has already been written under the
// hardened Go walk's containment checks.
//
// The system tar applies no containment of its own, so it is kept away from the
// install directory entirely: it extracts into an empty staging directory, and
// aqua then moves each member to its validated destination itself. Two properties
// are needed for that to be airtight, and both are enforced below:
//
//   - the system tar must not select any entry other than the sparse members
//     (validateSparseMember), otherwise it could plant a symlink pointing outside
//     the staging directory and write a later entry through it; and
//   - each member's destination must not resolve outside dest through a symlink
//     the walk planted (escapesDest, re-checked at the moment of the move).
//
// A tar binary on PATH is required; on macOS and Linux it is available by
// default. The system tar auto-detects the compression (gzip, xz, zstd, ...) from
// the archive content, so no compression flag is needed.
func (h *handler) extractSparseMembers(ctx context.Context, file string) error {
	for _, m := range h.sparseMembers {
		if err := h.validateSparseMember(m); err != nil {
			return err
		}
	}
	if h.executor == nil {
		return errors.New("cannot extract a GNU sparse archive: no command executor is available")
	}
	h.logger.Warn("the archive contains GNU sparse files unsupported by the Go tar reader; extracting them with the system tar command")

	// The staging directory is created inside dest so that it is on the same
	// filesystem, which keeps the move below a rename that preserves each file's
	// sparse allocation. It is removed again whether or not the extraction
	// succeeds.
	stagingDir, err := os.MkdirTemp(h.dest, ".aqua-sparse-")
	if err != nil {
		return fmt.Errorf("create a staging directory for GNU sparse files: %w", err)
	}
	defer os.RemoveAll(stagingDir)

	// "--" terminates option parsing so a member name starting with "-" is
	// treated as an operand, not a flag.
	args := []string{"-x", "-f", file, "-C", stagingDir, "--"}
	for _, m := range h.sparseMembers {
		args = append(args, m.name)
	}
	cmd := osexec.Command(ctx, "tar", args...)
	if _, err := h.executor.ExecAndOutputWhenFailure(cmd); err != nil {
		return fmt.Errorf("extract GNU sparse files with the system tar command (a `tar` binary on PATH is required for GNU sparse archives): %w", err)
	}

	for _, m := range h.sparseMembers {
		if err := h.moveSparseMember(stagingDir, m); err != nil {
			return err
		}
	}
	return nil
}

// tarMemberMetaChars are the characters a sparse member's name may not contain.
// bsdtar -- the tar of macOS, FreeBSD and Windows -- matches the member operands
// of an extraction as shell globs, so an entry named "*" would make it extract
// every member of the archive. Backslash is rejected too: it is the escape
// character of that glob syntax, and it is the one character normalizePath
// rewrites, so a name containing it would be validated at a different path from
// the one the system tar writes.
const tarMemberMetaChars = `*?[\`

// validateSparseMember reports whether m's name can be handed to the system tar
// as a member operand without selecting any other entry of the archive. tar
// matches an operand as a pattern rather than as a literal name: bsdtar
// glob-matches it, and every tar also selects the entries below "<name>/". A
// member that selected other entries would let the system tar write entries the
// Go walk never validated -- including a symlink pointing outside the staging
// directory followed by a regular file written through it, which escapes.
func (h *handler) validateSparseMember(m sparseMember) error {
	if strings.ContainsAny(m.name, tarMemberMetaChars) {
		return fmt.Errorf("%w: %s", errSparseMemberName, m.name)
	}
	name := cleanMemberName(m.name)
	selected := 0
	for _, entry := range h.entryNames {
		switch e := cleanMemberName(entry); {
		case e == name:
			// The member selects itself once; a second entry of the same name is
			// selected too, and the system tar would extract both.
			selected++
		case strings.HasPrefix(e, name+"/"):
			return fmt.Errorf("%w: %s also selects %s", errSparseMemberName, m.name, entry)
		}
	}
	if selected != 1 {
		return fmt.Errorf("%w: %s selects %d entries", errSparseMemberName, m.name, selected)
	}
	return nil
}

// cleanMemberName normalizes an entry name the way tar does before matching it,
// so that "./foo" and "foo" compare equal. Names in a tar archive are always
// slash-separated, hence path rather than filepath.
func cleanMemberName(name string) string {
	return path.Clean("/" + name)
}

// moveSparseMember moves one extracted sparse member out of the staging
// directory to its destination. os.Rename preserves the file's sparse allocation
// and, unlike an O_CREATE write, replaces a symlink standing at dstPath rather
// than following it.
func (h *handler) moveSparseMember(stagingDir string, m sparseMember) error {
	srcPath := filepath.Join(stagingDir, m.name)
	if !within(stagingDir, srcPath) {
		return fmt.Errorf("%w: %s", errEscapeDest, m.name)
	}
	// Re-validate the destination now that the walk has planted every symlink the
	// archive contains. A member whose destination resolves outside dest --
	// through a symlink planted by any entry, regardless of archive order -- must
	// not be moved there, because creating its parent directories would follow
	// that symlink out of dest.
	if h.escapesDest(m.dstPath) {
		return fmt.Errorf("%w: %s", errEscapeDest, m.name)
	}
	if err := osfile.MkdirAll(filepath.Dir(m.dstPath)); err != nil {
		return fmt.Errorf("create a directory for a GNU sparse file: %w", err)
	}
	if err := os.Rename(srcPath, m.dstPath); err != nil {
		return fmt.Errorf("move an extracted GNU sparse file to the extraction directory: %w", err)
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
	dst, err := os.Create(filepath.Join(h.dest, strings.TrimSuffix(h.filename, filepath.Ext(h.filename))))
	if err != nil {
		return fmt.Errorf("create a destination file: %w", err)
	}
	defer dst.Close()
	if _, err := io.Copy(dst, rc); err != nil {
		return fmt.Errorf("copy decompressed data: %w", err)
	}
	return nil
}
