package unarchive

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mholt/archiver/v3"
	"github.com/spf13/afero"
)

func Test_getUnarchiver(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		src   *File
		dest  string
		isErr bool
		exp   coreUnarchiver
	}{
		{
			title: "raw",
			src: &File{
				Type:     "raw",
				Filename: "foo",
			},
			dest: "/home/foo/.aqua/pkgs/github_release/github.com/aquaproj/foo/v1.0.0/foo",
			exp: &rawUnarchiver{
				dest: "/home/foo/.aqua/pkgs/github_release/github.com/aquaproj/foo/v1.0.0/foo/foo",
			},
		},
		{
			title: "ext is tar.gz",
			src: &File{
				Filename: "foo.tar.gz",
			},
			dest: "/home/foo/.aqua/pkgs/github_release/github.com/aquaproj/foo/v1.0.0/foo",
			exp: &unarchiverWithUnarchiver{
				dest:       "/home/foo/.aqua/pkgs/github_release/github.com/aquaproj/foo/v1.0.0/foo",
				unarchiver: archiver.NewTarGz(),
			},
		},
		{
			title: "ext is bz2",
			src: &File{
				Filename: "yoo.bz2",
			},
			dest: "/home/foo/.aqua/pkgs/github_release/github.com/aquaproj/foo/v1.0.0/foo",
			exp: &Decompressor{
				dest:         "/home/foo/.aqua/pkgs/github_release/github.com/aquaproj/foo/v1.0.0/foo/yoo",
				decompressor: archiver.NewBz2(),
			},
		},
		{
			title: "ext is dmg",
			src: &File{
				Type:     "dmg",
				Filename: "yoo.dmg",
			},
			dest: "/home/foo/.aqua/pkgs/github_release/github.com/aquaproj/foo/v1.0.0/foo",
			exp: &dmgUnarchiver{
				dest: "/home/foo/.aqua/pkgs/github_release/github.com/aquaproj/foo/v1.0.0/foo/yoo.dmg",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			unarchiver, err := getUnarchiver(d.src, d.dest)
			if d.isErr {
				if err == nil {
					t.Fatal("error should be returned")
				}
				return
			}
			if diff := cmp.Diff(d.exp, unarchiver, cmp.AllowUnexported(
				rawUnarchiver{}, unarchiverWithUnarchiver{}, Decompressor{}, dmgUnarchiver{}), cmpopts.IgnoreUnexported(archiver.TarGz{}, archiver.Tar{}, archiver.Bz2{}, afero.MemMapFs{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
