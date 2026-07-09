package unarchive_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/spf13/afero"
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

	fs := afero.NewOsFs()
	src := &unarchive.File{
		Filename: "malicious.tar.gz",
		Body:     download.NewDownloadedFile(fs, io.NopCloser(bytes.NewReader(archive)), nil),
	}
	if err := unarchive.New(nil, fs).Unarchive(ctx, logger, src, dest); err == nil {
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

	fs := afero.NewOsFs()
	src := &unarchive.File{
		Filename: "malicious.tar.gz",
		Body:     download.NewDownloadedFile(fs, io.NopCloser(bytes.NewReader(archive)), nil),
	}
	if err := unarchive.New(nil, fs).Unarchive(ctx, logger, src, dest); err == nil {
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

	fs := afero.NewOsFs()
	src := &unarchive.File{
		Filename: "malicious.tar.gz",
		Body:     download.NewDownloadedFile(fs, io.NopCloser(bytes.NewReader(archive)), nil),
	}
	if err := unarchive.New(nil, fs).Unarchive(ctx, logger, src, dest); err == nil {
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

	fs := afero.NewOsFs()
	src := &unarchive.File{
		Filename: "rootfs.tar.gz",
		Body:     download.NewDownloadedFile(fs, io.NopCloser(bytes.NewReader(archive)), nil),
	}
	if err := unarchive.New(nil, fs).Unarchive(ctx, logger, src, dest); err != nil {
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

	fs := afero.NewOsFs()
	src := &unarchive.File{
		Filename: "tool.tar.gz",
		Body:     download.NewDownloadedFile(fs, io.NopCloser(bytes.NewReader(archive)), nil),
	}
	if err := unarchive.New(nil, fs).Unarchive(ctx, logger, src, dest); err != nil {
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

	fs := afero.NewOsFs()
	src := &unarchive.File{
		Filename: "malicious.tar.gz",
		Body:     download.NewDownloadedFile(fs, io.NopCloser(bytes.NewReader(archive)), nil),
	}
	if err := unarchive.New(nil, fs).Unarchive(ctx, logger, src, dest); err == nil {
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
			fs := afero.NewOsFs()
			body, _, err := getReadCloser(ctx, httpDownloader, d.body, d.url)
			if body != nil {
				defer body.Close()
			}
			if err != nil {
				t.Fatal(err)
			}
			d.src.Body = download.NewDownloadedFile(fs, body, nil)
			unarchiver := unarchive.New(nil, fs)
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
