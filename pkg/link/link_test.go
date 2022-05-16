package link_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/link"
)

func TestMockLinker_Lstat(t *testing.T) {
	t.Parallel()
	data := []struct {
		name  string
		files map[string]string
		src   string
		isErr bool
	}{
		{
			name: "file found",
			files: map[string]string{
				"/home/foo/foo": "foo",
			},
			src: "foo",
		},
		{
			name:  "file isn't found",
			files: map[string]string{},
			src:   "foo",
			isErr: true,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			linker := link.NewMockLinker()
			for dest, src := range d.files {
				if err := linker.Symlink(dest, src); err != nil {
					t.Fatal(err)
				}
			}

			_, err := linker.Lstat(d.src)
			if err != nil {
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

func TestMockLinker_Readlink(t *testing.T) {
	t.Parallel()
	data := []struct {
		name  string
		files map[string]string
		src   string
		exp   string
		isErr bool
	}{
		{
			name: "file found",
			files: map[string]string{
				"/home/foo/foo": "foo",
			},
			src: "foo",
			exp: "/home/foo/foo",
		},
		{
			name:  "file isn't found",
			src:   "foo",
			isErr: true,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			linker := link.NewMockLinker()
			for dest, src := range d.files {
				if err := linker.Symlink(dest, src); err != nil {
					t.Fatal(err)
				}
			}

			dest, err := linker.Readlink(d.src)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if dest != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, dest)
			}
		})
	}
}
