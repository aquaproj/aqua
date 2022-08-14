package checksum_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/spf13/afero"
)

func TestChecksums_Get(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		m    map[string]string
		key  string
		exp  string
	}{
		{
			name: "key not found",
			key:  "foo",
			exp:  "",
		},
		{
			name: "key is found",
			key:  "foo",
			m: map[string]string{
				"foo": "bar",
			},
			exp: "bar",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			checksums := checksum.New()
			for k, v := range d.m {
				checksums.Set(k, v)
			}
			v := checksums.Get(d.key)
			if v != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, v)
			}
		})
	}
}

func TestChecksums_ReadFile(t *testing.T) {
	t.Parallel()
	data := []struct {
		name  string
		m     map[string]string
		p     string
		isErr bool
	}{
		{
			name: "file not found",
			p:    ".aqua-checksums.json",
		},
		{
			name: "file is found",
			m: map[string]string{
				".aqua-checksums.json": `{
  "github_release/github.com/cli/cli/v2.10.1/gh_2.10.1_macOS_amd64.tar.gz": "xxx"
}`,
			},
			p: ".aqua-checksums.json",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for k, v := range d.m {
				if err := afero.WriteFile(fs, k, []byte(v), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			checksums := checksum.New()
			if err := checksums.ReadFile(fs, d.p); err != nil {
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

func TestChecksums_UpdateFile(t *testing.T) {
	t.Parallel()
	data := []struct {
		name  string
		m     map[string]string
		p     string
		isErr bool
	}{
		{
			name: "normal",
			m: map[string]string{
				".aqua-checksums.json": `{
  "github_release/github.com/cli/cli/v2.10.1/gh_2.10.1_macOS_amd64.tar.gz": "xxx"
}`,
			},
			p: ".aqua-checksums.json",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			checksums := checksum.New()
			for k, v := range d.m {
				checksums.Set(k, v)
			}
			if err := checksums.UpdateFile(fs, d.p); err != nil {
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

func TestGetChecksumFilePathFromConfigFilePath(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name        string
		cfgFilePath string
		exp         string
		files       map[string]string
	}{
		{
			name:        "new",
			cfgFilePath: "aqua.yaml",
			exp:         "aqua-checksums.json",
		},
		{
			name:        "aqua-checksums.json > .aqua-checksums.json",
			cfgFilePath: "aqua.yaml",
			exp:         "aqua-checksums.json",
			files: map[string]string{
				"aqua-checksums.json":  "",
				".aqua-checksums.json": "",
			},
		},
		{
			name:        ".aqua-checksums.json",
			cfgFilePath: "aqua.yaml",
			exp:         ".aqua-checksums.json",
			files: map[string]string{
				".aqua-checksums.json": "",
			},
		},
		{
			name:        "new absolute",
			cfgFilePath: "/home/foo/aqua.yaml",
			exp:         "/home/foo/aqua-checksums.json",
		},
		{
			name:        "absolute aqua-checksums.json > .aqua-checksums.json",
			cfgFilePath: "/home/foo/aqua.yaml",
			exp:         "/home/foo/aqua-checksums.json",
			files: map[string]string{
				"/home/foo/.aqua-checksums.json": "",
				"/home/foo/aqua-checksums.json":  "",
			},
		},
		{
			name:        "absolute .aqua-checksums.json",
			cfgFilePath: "/home/foo/aqua.yaml",
			exp:         "/home/foo/.aqua-checksums.json",
			files: map[string]string{
				"/home/foo/.aqua-checksums.json": "",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for k, v := range d.files {
				if err := afero.WriteFile(fs, k, []byte(v), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			p, err := checksum.GetChecksumFilePathFromConfigFilePath(fs, d.cfgFilePath)
			if err != nil {
				t.Fatal(err)
			}
			if p != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, p)
			}
		})
	}
}
