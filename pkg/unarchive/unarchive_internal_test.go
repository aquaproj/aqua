package unarchive

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mholt/archiver/v3"
)

func Test_getUnarchiver(t *testing.T) {
	t.Parallel()
	data := []struct {
		title    string
		filename string
		typ      string
		dest     string
		isErr    bool
		exp      Unarchiver
	}{
		{
			title:    "raw",
			typ:      "raw",
			filename: "foo",
			dest:     "/home/foo/.aqua/pkgs/github_release/github.com/aquaproj/foo/v1.0.0/foo",
			exp: &rawUnarchiver{
				dest: "/home/foo/.aqua/pkgs/github_release/github.com/aquaproj/foo/v1.0.0/foo/foo",
			},
		},
		{
			title:    "ext is tar.gz",
			filename: "foo.tar.gz",
			dest:     "/home/foo/.aqua/pkgs/github_release/github.com/aquaproj/foo/v1.0.0/foo",
			exp: &unarchiverWithUnarchiver{
				dest:       "/home/foo/.aqua/pkgs/github_release/github.com/aquaproj/foo/v1.0.0/foo",
				unarchiver: archiver.NewTarGz(),
			},
		},
		{
			title:    "ext is bz2",
			filename: "yoo.bz2",
			dest:     "/home/foo/.aqua/pkgs/github_release/github.com/aquaproj/foo/v1.0.0/foo",
			exp: &Decompressor{
				dest:         "/home/foo/.aqua/pkgs/github_release/github.com/aquaproj/foo/v1.0.0/foo/yoo",
				decompressor: archiver.NewBz2(),
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			unarchiver, err := getUnarchiver(d.filename, d.typ, d.dest)
			if d.isErr {
				if err == nil {
					t.Fatal("error should be returned")
				}
				return
			}
			if diff := cmp.Diff(d.exp, unarchiver, cmp.AllowUnexported(
				rawUnarchiver{}, unarchiverWithUnarchiver{}, Decompressor{}), cmpopts.IgnoreUnexported(archiver.TarGz{}, archiver.Tar{}, archiver.Bz2{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

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
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			f := IsUnarchived(d.archiveType, d.assetName)
			if f != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, f)
			}
		})
	}
}
