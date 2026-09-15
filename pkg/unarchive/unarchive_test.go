package unarchive_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
)

func TestIsUnarchived(t *testing.T) {
	t.Parallel()
	data := []struct {
		title       string
		archiveType string
		assetName   string
		exp         bool
	}{
		{
			title:     "tar.gz",
			assetName: "foo.tar.gz",
			exp:       false,
		},
		{
			title:     "archive is empty and assetName has no extension",
			assetName: "foo",
			exp:       true,
		},
		{
			title:       "archiveType is raw",
			assetName:   "foo-v3.0.0",
			archiveType: "raw",
			exp:         true,
		},
		{
			title:       "archiveType is set and isn't raw",
			assetName:   "foo",
			archiveType: "tar.gz",
			exp:         false,
		},
		{
			title:     ".exe is raw",
			assetName: "foo.exe",
			exp:       true,
		},
		{
			title:     ".dmg",
			assetName: "foo.dmg",
			exp:       false,
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			f := unarchive.IsUnarchived(d.archiveType, d.assetName)
			if f != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, f)
			}
		})
	}
}

type tarEntry struct {
	hdr     *tar.Header
	payload []byte
}

// buildTarGz builds a gzip-compressed tar archive from the given entries. When
// an entry has a payload, its header Size is set from the payload length.
func buildTarGz(t *testing.T, entries ...tarEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	for _, e := range entries {
		e.hdr.Size = int64(len(e.payload))
		if err := tw.WriteHeader(e.hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(e.payload); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// TestUnarchiver_Unarchive_symlinkTraversal verifies that an archive cannot use
// a symlink pointing outside the extraction directory followed by a regular file
// entry at the same path to write outside the destination.
// See https://github.com/aquaproj/aqua/security/advisories/GHSA-mf5c-hw34-4hpp
func TestUnarchiver_Unarchive_symlinkTraversal(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	logger := slog.New(slog.DiscardHandler)

	dest := t.TempDir()
	outsideDir := t.TempDir()
	outside := filepath.Join(outsideDir, "outside-target")
	if err := os.WriteFile(outside, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}

	// A symlink "pwn" -> outside target, then a regular file "pwn" whose write
	// would follow the planted symlink.
	archive := buildTarGz(
		t,
		tarEntry{hdr: &tar.Header{Name: "pwn", Typeflag: tar.TypeSymlink, Linkname: outside, Mode: 0o777}},
		tarEntry{hdr: &tar.Header{Name: "pwn", Typeflag: tar.TypeReg, Mode: 0o644}, payload: []byte("PWNED_BY_AQUA_SYMLINK_TRAVERSAL")},
	)

	src := &unarchive.File{
		Filename: "malicious.tar.gz",
		Body:     download.NewDownloadedFile(io.NopCloser(bytes.NewReader(archive)), nil),
	}
	if err := unarchive.New(nil).Unarchive(ctx, logger, src, dest); err == nil {
		t.Fatal("an error must be returned for a regular file that follows a symlink escaping the extraction directory")
	}

	got, err := os.ReadFile(outside)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "original" {
		t.Fatalf("the outside file was modified through a symlink: %q", got)
	}
}

// TestUnarchiver_Unarchive_symlinkDirTraversal verifies that a regular file
// whose path descends through a symlink pointing outside the extraction
// directory cannot be written outside it. This is the multi-entry variant of
// the traversal: "pwn" -> outside dir, then "pwn/file".
// See https://github.com/aquaproj/aqua/security/advisories/GHSA-mf5c-hw34-4hpp
func TestUnarchiver_Unarchive_symlinkDirTraversal(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	logger := slog.New(slog.DiscardHandler)

	dest := t.TempDir()
	outsideDir := t.TempDir()

	// A symlink "pwn" -> outside dir, then a regular file "pwn/file" whose write
	// would descend through the planted symlink.
	archive := buildTarGz(
		t,
		tarEntry{hdr: &tar.Header{Name: "pwn", Typeflag: tar.TypeSymlink, Linkname: outsideDir, Mode: 0o777}},
		tarEntry{hdr: &tar.Header{Name: "pwn/file", Typeflag: tar.TypeReg, Mode: 0o644}, payload: []byte("PWNED_BY_AQUA_SYMLINK_DIR_TRAVERSAL")},
	)

	src := &unarchive.File{
		Filename: "malicious.tar.gz",
		Body:     download.NewDownloadedFile(io.NopCloser(bytes.NewReader(archive)), nil),
	}
	if err := unarchive.New(nil).Unarchive(ctx, logger, src, dest); err == nil {
		t.Fatal("an error must be returned for a file escaping via a symlinked directory")
	}
	if _, err := os.Stat(filepath.Join(outsideDir, "file")); !os.IsNotExist(err) {
		t.Fatalf("a file was written outside dest through a symlinked directory: %v", err)
	}
}

// TestUnarchiver_Unarchive_danglingSymlinkTraversal verifies that a symlink
// whose target does not exist yet, followed by a regular file at the same path,
// cannot create a file outside the extraction directory. filepath.EvalSymlinks
// fails on such a dangling link, so the escape guard must not rely on the final
// target already existing.
// See https://github.com/aquaproj/aqua/security/advisories/GHSA-mf5c-hw34-4hpp
func TestUnarchiver_Unarchive_danglingSymlinkTraversal(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	logger := slog.New(slog.DiscardHandler)

	dest := t.TempDir()
	outsideDir := t.TempDir()
	// The target intentionally does not exist yet.
	outside := filepath.Join(outsideDir, "created-outside-target")

	// A symlink "pwn" -> not-yet-existing outside target, then a regular file
	// "pwn" whose O_CREATE write would follow the dangling symlink.
	archive := buildTarGz(
		t,
		tarEntry{hdr: &tar.Header{Name: "pwn", Typeflag: tar.TypeSymlink, Linkname: outside, Mode: 0o777}},
		tarEntry{hdr: &tar.Header{Name: "pwn", Typeflag: tar.TypeReg, Mode: 0o644}, payload: []byte("PWNED_BY_AQUA_DANGLING_SYMLINK")},
	)

	src := &unarchive.File{
		Filename: "malicious.tar.gz",
		Body:     download.NewDownloadedFile(io.NopCloser(bytes.NewReader(archive)), nil),
	}
	if err := unarchive.New(nil).Unarchive(ctx, logger, src, dest); err == nil {
		t.Fatal("an error must be returned for a regular file that follows a dangling symlink escaping the extraction directory")
	}
	if _, err := os.Stat(outside); !os.IsNotExist(err) {
		t.Fatalf("a file was created outside dest through a dangling symlink: %v", err)
	}
}

// TestUnarchiver_Unarchive_symlinkOutsideAllowed verifies that a symlink whose
// target points outside the extraction directory is created successfully when
// no later entry follows it to write outside. Such symlinks are common in
// legitimate archives such as root filesystem images (e.g. "var/run -> /run"),
// so extraction must not fail on their mere presence.
func TestUnarchiver_Unarchive_symlinkOutsideAllowed(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	logger := slog.New(slog.DiscardHandler)

	dest := t.TempDir()

	// An absolute symlink escaping dest (mirroring a rootfs "var/run -> /run")
	// followed by a normal file that must still be extracted.
	archive := buildTarGz(
		t,
		tarEntry{hdr: &tar.Header{Name: "var/run", Typeflag: tar.TypeSymlink, Linkname: "/run", Mode: 0o777}},
		tarEntry{hdr: &tar.Header{Name: "bin/tool", Typeflag: tar.TypeReg, Mode: 0o755}, payload: []byte("hello")},
	)

	src := &unarchive.File{
		Filename: "rootfs.tar.gz",
		Body:     download.NewDownloadedFile(io.NopCloser(bytes.NewReader(archive)), nil),
	}
	if err := unarchive.New(nil).Unarchive(ctx, logger, src, dest); err != nil {
		t.Fatalf("extraction must succeed for a benign escaping symlink: %v", err)
	}

	target, err := os.Readlink(filepath.Join(dest, "var/run"))
	if err != nil {
		t.Fatalf("the escaping symlink must be created: %v", err)
	}
	if target != "/run" {
		t.Fatalf("unexpected symlink target: %q", target)
	}
	got, err := os.ReadFile(filepath.Join(dest, "bin/tool"))
	if err != nil {
		t.Fatalf("the regular file must be extracted: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("unexpected file content: %q", got)
	}
}

// TestUnarchiver_Unarchive_rootEntry verifies that an archive whose first entry
// is the root directory "./" (as produced by e.g. crate-ci/typos releases)
// extracts successfully. Its dstPath equals dest, so its parent is legitimately
// outside dest and must not be treated as an escape.
func TestUnarchiver_Unarchive_rootEntry(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	logger := slog.New(slog.DiscardHandler)

	dest := t.TempDir()

	archive := buildTarGz(
		t,
		tarEntry{hdr: &tar.Header{Name: "./", Typeflag: tar.TypeDir, Mode: 0o755}},
		tarEntry{hdr: &tar.Header{Name: "./tool", Typeflag: tar.TypeReg, Mode: 0o755}, payload: []byte("hello")},
	)

	src := &unarchive.File{
		Filename: "tool.tar.gz",
		Body:     download.NewDownloadedFile(io.NopCloser(bytes.NewReader(archive)), nil),
	}
	if err := unarchive.New(nil).Unarchive(ctx, logger, src, dest); err != nil {
		t.Fatalf("extraction must succeed for an archive with a \"./\" root entry: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dest, "tool"))
	if err != nil {
		t.Fatalf("the regular file must be extracted: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("unexpected file content: %q", got)
	}
}

// TestUnarchiver_Unarchive_pathTraversal verifies that an archive entry whose
// path contains ".." cannot write outside the extraction directory.
// See https://github.com/aquaproj/aqua/security/advisories/GHSA-mf5c-hw34-4hpp
func TestUnarchiver_Unarchive_pathTraversal(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	logger := slog.New(slog.DiscardHandler)

	parent := t.TempDir()
	dest := filepath.Join(parent, "dest")
	outside := filepath.Join(parent, "outside-target")
	if err := os.WriteFile(outside, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}

	// A single regular file whose name escapes dest via "..".
	archive := buildTarGz(
		t,
		tarEntry{hdr: &tar.Header{Name: "../outside-target", Typeflag: tar.TypeReg, Mode: 0o644}, payload: []byte("PWNED_BY_AQUA_PATH_TRAVERSAL")},
	)

	src := &unarchive.File{
		Filename: "malicious.tar.gz",
		Body:     download.NewDownloadedFile(io.NopCloser(bytes.NewReader(archive)), nil),
	}
	if err := unarchive.New(nil).Unarchive(ctx, logger, src, dest); err == nil {
		t.Fatal("an error must be returned for an entry escaping the extraction directory")
	}

	got, err := os.ReadFile(outside)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "original" {
		t.Fatalf("the outside file was modified via path traversal: %q", got)
	}
}

// pkgutilStub stands in for the pkgutil command. It records whether the
// destination existed when the command was invoked, then creates it as
// "pkgutil --expand-full" does.
type pkgutilStub struct {
	args        []string
	destExisted bool
}

func (e *pkgutilStub) ExecAndOutputWhenFailure(cmd *osexec.Cmd) (int, error) {
	e.args = cmd.Args
	dest := cmd.Args[len(cmd.Args)-1]
	if _, err := os.Stat(dest); err == nil {
		e.destExisted = true
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return 1, fmt.Errorf("create the destination: %w", err)
	}
	return 0, nil
}

// newPkgFile returns a pkg format source file along with a destination
// directory that the caller has already created, as Installer.unarchive does.
func newPkgFile(t *testing.T) (*unarchive.File, string) {
	t.Helper()
	dest := filepath.Join(t.TempDir(), "dest")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	return &unarchive.File{
		Filename: "s3deploy_2.16.0_darwin-universal.pkg",
		Type:     unarchive.FormatPKG,
		Body:     download.NewDownloadedFile(io.NopCloser(strings.NewReader("pkg")), nil),
	}, dest
}

// TestUnarchiver_Unarchive_pkg verifies that the destination handed to
// "pkgutil --expand-full" does not exist when the command runs. pkgutil refuses
// to expand into an existing path and fails with "File exists", so every pkg
// format package failed to install once the caller started creating the
// extraction directory itself.
// See https://github.com/aquaproj/aqua/issues/5040
func TestUnarchiver_Unarchive_pkg(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	logger := slog.New(slog.DiscardHandler)

	src, dest := newPkgFile(t)
	executor := &pkgutilStub{}

	if err := unarchive.New(executor).Unarchive(ctx, logger, src, dest); err != nil {
		t.Fatal(err)
	}

	if executor.destExisted {
		t.Fatal("the destination must not exist when pkgutil --expand-full is executed")
	}
	if len(executor.args) != 4 || executor.args[1] != "--expand-full" || executor.args[3] != dest {
		t.Fatalf("unexpected command: %v", executor.args)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Fatal("the destination must be created by pkgutil")
	}
}

// A destination that already holds files belongs to something else, so it must
// not be removed to make room for pkgutil.
func TestUnarchiver_Unarchive_pkg_destNotEmpty(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	logger := slog.New(slog.DiscardHandler)

	src, dest := newPkgFile(t)
	payload := filepath.Join(dest, "Payload")
	if err := os.WriteFile(payload, []byte("installed by someone else"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := unarchive.New(&pkgutilStub{}).Unarchive(ctx, logger, src, dest); err == nil {
		t.Fatal("an error must be returned for a destination that is not empty")
	}

	b, err := os.ReadFile(payload)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "installed by someone else" {
		t.Fatalf("the existing file is %q, want %q", b, "installed by someone else")
	}
}

func getReadCloser(ctx context.Context, downloader download.HTTPDownloader, body, url string) (io.ReadCloser, int64, error) {
	if body != "" {
		return io.NopCloser(strings.NewReader(body)), 0, nil
	}
	return downloader.Download(ctx, url) //nolint:wrapcheck
}

func TestUnarchiver_Unarchive(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		src   *unarchive.File
		body  string
		url   string
		isErr bool
	}{
		{
			title: "raw",
			src: &unarchive.File{
				Filename: "aqua-installer",
			},
			body: `foo`,
		},
		{
			title: "tarball",
			url:   "https://github.com/suzuki-shunsuke/archives-for-test/raw/main/README.md.tar.gz",
			src: &unarchive.File{
				Filename: "README.md.tar.gz",
			},
		},
		{
			title: "decompressor",
			url:   "https://github.com/suzuki-shunsuke/archives-for-test/raw/main/README.md.bz2",
			src: &unarchive.File{
				Filename: "README.md.bz2",
			},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	httpDownloader := download.NewHTTPDownloader(logger, http.DefaultClient)
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			body, _, err := getReadCloser(ctx, httpDownloader, d.body, d.url)
			if body != nil {
				defer body.Close()
			}
			if err != nil {
				t.Fatal(err)
			}
			d.src.Body = download.NewDownloadedFile(body, nil)
			unarchiver := unarchive.New(nil)
			if err := unarchiver.Unarchive(ctx, logger, d.src, t.TempDir()); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}

// TestUnarchiver_Unarchive_gnuSparse verifies that an archive containing a GNU
// sparse file (PAX GNU.sparse.* records), which Go's archive/tar cannot reliably
// extract, is extracted by falling back to the system tar command.
// testdata/gnu-sparse.tar.gz holds a well-formed GNU sparse 1.0 member
// "sparse.img" of logical size 131072 with "HELLO" at offset 0 and a hole after.
func TestUnarchiver_Unarchive_gnuSparse(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("tar"); err != nil {
		t.Skip("the system tar command is required to extract GNU sparse archives")
	}
	ctx := t.Context()
	logger := slog.New(slog.DiscardHandler)

	archive, err := os.ReadFile(filepath.Join("testdata", "gnu-sparse.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}

	dest := t.TempDir()
	src := &unarchive.File{
		Filename: "gnu-sparse.tar.gz",
		Body:     download.NewDownloadedFile(io.NopCloser(bytes.NewReader(archive)), nil),
	}
	if err := unarchive.New(osexec.New()).Unarchive(ctx, logger, src, dest); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(dest, "sparse.img"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 131072 {
		t.Fatalf("extracted file size: want 131072, got %d", len(got))
	}
	if string(got[:5]) != "HELLO" {
		t.Fatalf("extracted file head: want HELLO, got %q", got[:5])
	}
	if got[70000] != 0 {
		t.Fatalf("the hole region must be zero, got %d at offset 70000", got[70000])
	}
}

// TestUnarchiver_Unarchive_gnuSparse_symlinkTraversal verifies that a GNU sparse
// member cannot be used to bypass the symlink-traversal guards. Extracting a
// sparse member shells out to the system tar, which applies no containment of its
// own, so a sparse member must neither disable aqua's checks on the archive's
// other entries nor let the system tar write through an escaping symlink.
// The archive plants a symlink whose backslash name aqua rewrites to "x/evil"
// pointing outside dest, then a sparse member reaching the system-tar fallback,
// then a regular file "x/evil/pwned" whose write would follow the planted
// symlink -- the layout of the reported bypass.
// See https://github.com/aquaproj/aqua/security/advisories/GHSA-286g-rf2x-cv99
func TestUnarchiver_Unarchive_gnuSparse_symlinkTraversal(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("tar"); err != nil {
		t.Skip("the system tar command is required to extract GNU sparse archives")
	}
	ctx := t.Context()
	logger := slog.New(slog.DiscardHandler)

	dest := t.TempDir()
	outsideDir := t.TempDir()
	outside := filepath.Join(outsideDir, "outside-target")
	if err := os.WriteFile(outside, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}

	archive := buildSparseTarGz(
		t,
		[]tarEntry{{hdr: &tar.Header{Name: `x\evil`, Typeflag: tar.TypeSymlink, Linkname: outside, Mode: 0o777}}},
		[]tarEntry{{hdr: &tar.Header{Name: "x/evil/pwned", Typeflag: tar.TypeReg, Mode: 0o644}, payload: []byte("PWNED_BY_AQUA_SPARSE_FALLBACK")}},
	)

	src := &unarchive.File{
		Filename: "malicious.tar.gz",
		Body:     download.NewDownloadedFile(io.NopCloser(bytes.NewReader(archive)), nil),
	}
	if err := unarchive.New(osexec.New()).Unarchive(ctx, logger, src, dest); err == nil {
		t.Fatal("an error must be returned for an archive that escapes the extraction directory through the sparse fallback")
	}

	got, err := os.ReadFile(outside)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "original" {
		t.Fatalf("the outside file was modified through the sparse fallback: %q", got)
	}
}

// buildSparseTarGz builds a gzip-compressed tar archive containing the leading
// entries, then the real GNU sparse member from testdata (so the extraction
// triggers aqua's system-tar fallback), then the trailing entries. It lets a
// test reproduce an archive that mixes attacker-controlled entries with a GNU
// sparse member that Go's archive/tar cannot synthesize.
func buildSparseTarGz(t *testing.T, leading, trailing []tarEntry) []byte {
	t.Helper()
	var raw bytes.Buffer
	for _, e := range leading {
		raw.Write(tarEntryBlocks(t, e))
	}
	raw.Write(sparseMemberBlocks(t))
	for _, e := range trailing {
		raw.Write(tarEntryBlocks(t, e))
	}
	return gzipTarBlocks(t, raw.Bytes())
}

// gzipTarBlocks terminates a stream of raw tar blocks with the two zero blocks
// that end an archive and gzips the result.
func gzipTarBlocks(t *testing.T, raw []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(raw); err != nil {
		t.Fatal(err)
	}
	if _, err := gw.Write(make([]byte, 1024)); err != nil { // two zero blocks terminate the archive
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// tarEntryBlocks returns the raw tar blocks (header plus padded payload) for a
// single entry, with the trailing zero-block trailer that tar.Writer.Close
// appends removed so the blocks can be spliced next to other entries.
func tarEntryBlocks(t *testing.T, e tarEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	e.hdr.Size = int64(len(e.payload))
	if err := tw.WriteHeader(e.hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(e.payload); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()[:buf.Len()-1024]
}

// sparseMemberBlocks returns the raw tar blocks of the GNU sparse member
// "sparse.img" stored in testdata/gnu-sparse.tar.gz, with the archive's trailing
// zero blocks removed so the member can be spliced between other entries.
func sparseMemberBlocks(t *testing.T) []byte {
	t.Helper()
	gzb, err := os.ReadFile(filepath.Join("testdata", "gnu-sparse.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	gr, err := gzip.NewReader(bytes.NewReader(gzb))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(gr)
	if err != nil {
		t.Fatal(err)
	}
	for len(raw) >= 512 {
		block := raw[len(raw)-512:]
		zero := true
		for _, c := range block {
			if c != 0 {
				zero = false
				break
			}
		}
		if !zero {
			break
		}
		raw = raw[:len(raw)-512]
	}
	return raw
}

// TestUnarchiver_Unarchive_gnuSparse_globMemberName verifies that a GNU sparse
// member whose name is a glob pattern is rejected instead of being handed to the
// system tar as a member operand. bsdtar -- the tar of macOS, FreeBSD and Windows
// -- matches those operands as shell globs, so a member named "*" would make it
// extract every member of the archive, i.e. entries the Go walk never validated
// at the path the system tar writes them to.
func TestUnarchiver_Unarchive_gnuSparse_globMemberName(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("tar"); err != nil {
		t.Skip("the system tar command is required to extract GNU sparse archives")
	}
	dest := t.TempDir()

	var raw bytes.Buffer
	// aqua's Go walk writes this entry to "a/b" because normalizePath rewrites the
	// backslash; the system tar writes it to the literal name `a\b` instead.
	raw.Write(tarEntryBlocks(t, tarEntry{
		hdr:     &tar.Header{Name: `a\b`, Typeflag: tar.TypeReg, Mode: 0o644},
		payload: []byte("written-by-the-system-tar"),
	}))
	raw.Write(renamedSparseMemberBlocks(t, "*"))

	src := &unarchive.File{
		Filename: "malicious.tar.gz",
		Body:     download.NewDownloadedFile(io.NopCloser(bytes.NewReader(gzipTarBlocks(t, raw.Bytes()))), nil),
	}
	if err := unarchive.New(osexec.New()).Unarchive(t.Context(), slog.New(slog.DiscardHandler), src, dest); err == nil {
		t.Fatal("an error must be returned for a GNU sparse member whose name is a glob pattern")
	}
	if _, err := os.Lstat(filepath.Join(dest, `a\b`)); err == nil {
		t.Fatal(`the system tar extracted "a\b", an entry the Go walk never validated at that path`)
	}
}

// TestUnarchiver_Unarchive_gnuSparse_prefixMemberName verifies that a GNU sparse
// member is rejected when the archive also holds entries below "<name>/". Every
// tar selects those entries along with the named member, which hands the system
// tar -- which applies no containment of its own -- entries the Go walk never
// wrote. The archive below is the escape this buys: the Go walk completes with
// warnings only, because the symlink entry cannot be created over the regular
// file of the same name, so the write through it lands on a plain file inside
// dest. A system tar replaying those three entries instead unlinks the regular
// file, plants the symlink pointing outside dest, and writes through it.
func TestUnarchiver_Unarchive_gnuSparse_prefixMemberName(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("tar"); err != nil {
		t.Skip("the system tar command is required to extract GNU sparse archives")
	}
	dest := t.TempDir()
	outsideDir := t.TempDir()

	var raw bytes.Buffer
	for _, e := range []tarEntry{
		{hdr: &tar.Header{Name: "x/link", Typeflag: tar.TypeReg, Mode: 0o644}, payload: []byte("decoy")},
		{hdr: &tar.Header{Name: "x/link", Typeflag: tar.TypeSymlink, Linkname: outsideDir, Mode: 0o777}},
		{hdr: &tar.Header{Name: "x/link/pwned", Typeflag: tar.TypeReg, Mode: 0o644}, payload: []byte("PWNED_BY_AQUA_SPARSE_MEMBER_PREFIX")},
	} {
		raw.Write(tarEntryBlocks(t, e))
	}
	raw.Write(renamedSparseMemberBlocks(t, "x"))

	src := &unarchive.File{
		Filename: "malicious.tar.gz",
		Body:     download.NewDownloadedFile(io.NopCloser(bytes.NewReader(gzipTarBlocks(t, raw.Bytes()))), nil),
	}
	if err := unarchive.New(osexec.New()).Unarchive(t.Context(), slog.New(slog.DiscardHandler), src, dest); err == nil {
		t.Fatal("an error must be returned for a GNU sparse member that also selects the entries below it")
	}

	if _, err := os.Lstat(filepath.Join(outsideDir, "pwned")); err == nil {
		t.Fatal("a file was written outside the extraction directory")
	}
	// The system tar must not have replaced the regular file the Go walk wrote
	// with the escaping symlink: planting it is the first half of the escape, and
	// whether the write through it then succeeds is up to the host tar.
	fi, err := os.Lstat(filepath.Join(dest, "x", "link"))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		t.Fatal("the system tar planted a symlink pointing outside the extraction directory")
	}
}

// TestUnarchiver_Unarchive_gnuSparse_subDir verifies that a GNU sparse member
// stored in a subdirectory is moved out of the staging directory to the right
// place, and that the staging directory itself does not survive the extraction.
func TestUnarchiver_Unarchive_gnuSparse_subDir(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("tar"); err != nil {
		t.Skip("the system tar command is required to extract GNU sparse archives")
	}
	dest := t.TempDir()

	src := &unarchive.File{
		Filename: "gnu-sparse.tar.gz",
		Body: download.NewDownloadedFile(
			io.NopCloser(bytes.NewReader(gzipTarBlocks(t, renamedSparseMemberBlocks(t, "d/sparse.img")))), nil),
	}
	if err := unarchive.New(osexec.New()).Unarchive(t.Context(), slog.New(slog.DiscardHandler), src, dest); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(dest, "d", "sparse.img"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 131072 || string(got[:5]) != "HELLO" {
		t.Fatalf("extracted file: want 131072 bytes starting with HELLO, got %d bytes starting with %q", len(got), got[:5])
	}
	entries, err := os.ReadDir(dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "d" {
		t.Fatalf(`the extraction directory must hold only "d", got %v`, entries)
	}
}

// renamedSparseMemberBlocks returns the raw tar blocks of the GNU sparse member
// stored in testdata/gnu-sparse.tar.gz with its GNU.sparse.name PAX record
// rewritten to name, so a test can control the member name aqua passes to the
// system tar. Go's archive/tar cannot write sparse members, hence the rewrite of
// a real one.
func renamedSparseMemberBlocks(t *testing.T, name string) []byte {
	t.Helper()
	raw := sparseMemberBlocks(t)
	if len(raw) < 1024 {
		t.Fatalf("the sparse member must hold a PAX header block and its records, got %d bytes", len(raw))
	}
	records := rewritePAXRecord(t, raw[512:1024], "GNU.sparse.name", name)
	if len(records) > 512 {
		t.Fatalf("the rewritten PAX records must fit in one block, got %d bytes", len(records))
	}

	out := make([]byte, 0, len(raw))
	hdr := make([]byte, 512)
	copy(hdr, raw[:512])
	// The PAX extended header's payload size changed, so rewrite the size field
	// (12 bytes at offset 124) and the header checksum.
	copy(hdr[124:136], fmt.Sprintf("%011o\x00", len(records)))
	setTarChecksum(hdr)
	out = append(out, hdr...)
	block := make([]byte, 512)
	copy(block, records)
	out = append(out, block...)
	return append(out, raw[1024:]...)
}

// rewritePAXRecord returns the PAX extended header records in data with key's
// value replaced by value, keeping every other record untouched.
func rewritePAXRecord(t *testing.T, data []byte, key, value string) []byte {
	t.Helper()
	var out strings.Builder
	for len(data) > 0 && data[0] != 0 {
		sp := bytes.IndexByte(data, ' ')
		if sp < 0 {
			t.Fatalf("a PAX record must start with its length, got %q", data)
		}
		size, err := strconv.Atoi(string(data[:sp]))
		if err != nil || size > len(data) {
			t.Fatalf("invalid PAX record length %q: %v", data[:sp], err)
		}
		record := string(data[:size])
		if k, _, ok := strings.Cut(record[sp+1:], "="); ok && k == key {
			record = paxRecord(key, value)
		}
		out.WriteString(record)
		data = data[size:]
	}
	return []byte(out.String())
}

// paxRecord formats a PAX extended header record, whose length prefix counts the
// length prefix itself.
func paxRecord(key, value string) string {
	body := " " + key + "=" + value + "\n"
	for digits := 1; ; digits++ {
		record := strconv.Itoa(len(body)+digits) + body
		if len(record) == len(body)+digits {
			return record
		}
	}
}

// setTarChecksum stores the checksum of a 512 byte tar header block in its
// chksum field, which is computed with that field filled with spaces.
func setTarChecksum(hdr []byte) {
	for i := 148; i < 156; i++ {
		hdr[i] = ' '
	}
	sum := 0
	for _, c := range hdr {
		sum += int(c)
	}
	copy(hdr[148:], fmt.Sprintf("%06o\x00 ", sum))
}
