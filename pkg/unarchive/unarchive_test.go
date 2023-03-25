package unarchive_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/unarchive"
	"github.com/sirupsen/logrus"
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
			title:       "archiveTyps is set and isn't raw",
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
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			f := unarchive.IsUnarchived(d.archiveType, d.assetName)
			if f != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, f)
			}
		})
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
	ctx := context.Background()
	httpDownloader := download.NewHTTPDownloader(http.DefaultClient)
	unarchiver := &unarchive.UnarchiverImpl{}
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewOsFs()
			body, _, err := getReadCloser(ctx, httpDownloader, d.body, d.url)
			if body != nil {
				defer body.Close()
			}
			if err != nil {
				t.Fatal(err)
			}
			d.src.Body = body
			if err := unarchiver.Unarchive(ctx, logE, d.src, t.TempDir(), fs, nil); err != nil {
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
