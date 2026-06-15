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

	// Build a tar.gz: a symlink "pwn" -> outside target, then a regular file
	// "pwn" whose write would follow the planted symlink.
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	if err := tw.WriteHeader(&tar.Header{
		Name:     "pwn",
		Typeflag: tar.TypeSymlink,
		Linkname: outside,
		Mode:     0o777,
	}); err != nil {
		t.Fatal(err)
	}
	payload := []byte("PWNED_BY_AQUA_SYMLINK_TRAVERSAL")
	if err := tw.WriteHeader(&tar.Header{
		Name:     "pwn",
		Typeflag: tar.TypeReg,
		Mode:     0o644,
		Size:     int64(len(payload)),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}

	fs := afero.NewOsFs()
	src := &unarchive.File{
		Filename: "malicious.tar.gz",
		Body:     download.NewDownloadedFile(fs, io.NopCloser(bytes.NewReader(buf.Bytes())), nil),
	}
	if err := unarchive.New(nil, fs).Unarchive(ctx, logger, src, dest); err == nil {
		t.Fatal("an error must be returned for a symlink escaping the extraction directory")
	}

	got, err := os.ReadFile(outside)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "original" {
		t.Fatalf("the outside file was modified through a symlink: %q", got)
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

	// Build a tar.gz with a single regular file whose name escapes dest via "..".
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	payload := []byte("PWNED_BY_AQUA_PATH_TRAVERSAL")
	if err := tw.WriteHeader(&tar.Header{
		Name:     "../outside-target",
		Typeflag: tar.TypeReg,
		Mode:     0o644,
		Size:     int64(len(payload)),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}

	fs := afero.NewOsFs()
	src := &unarchive.File{
		Filename: "malicious.tar.gz",
		Body:     download.NewDownloadedFile(fs, io.NopCloser(bytes.NewReader(buf.Bytes())), nil),
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
