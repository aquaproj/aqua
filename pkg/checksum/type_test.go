package checksum_test

import (
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/google/go-cmp/cmp"
)

func TestChecksums_Get(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		m    map[string]*checksum.Checksum
		key  string
		exp  *checksum.Checksum
	}{
		{
			name: "key not found",
			key:  pkgFoo,
			exp:  nil,
		},
		{
			name: "key is found",
			key:  pkgFoo,
			m: map[string]*checksum.Checksum{
				pkgFoo: {
					ID:        pkgFoo,
					Checksum:  "bar",
					Algorithm: algoSHA256,
				},
			},
			exp: &checksum.Checksum{
				ID:        pkgFoo,
				Checksum:  "BAR",
				Algorithm: algoSHA256,
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			checksums := checksum.New()
			for k, v := range d.m {
				checksums.Set(k, v)
			}
			v := checksums.Get(d.key)
			if diff := cmp.Diff(v, d.exp); diff != "" {
				t.Fatal(diff)
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
			p:    fileDotAquaChecksums,
		},
		{
			name: "file is found",
			m: map[string]string{
				fileDotAquaChecksums: `{
  "github_release/github.com/cli/cli/v2.10.1/gh_2.10.1_macOS_amd64.tar.gz": "xxx"
}`,
			},
			p: fileDotAquaChecksums,
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			testutil.WriteFiles(t, dir, d.m)
			checksums := checksum.New()
			if err := checksums.ReadFile(filepath.Join(dir, d.p)); err != nil {
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
		m     []*checksum.Checksum
		p     string
		isErr bool
	}{
		{
			name: "normal",
			m: []*checksum.Checksum{
				{
					ID:        "github_release/github.com/cli/cli/v2.10.1/gh_2.10.1_macOS_amd64.tar.gz",
					Checksum:  "xxx",
					Algorithm: algoSHA256,
				},
			},
			p: fileDotAquaChecksums,
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			p := filepath.Join(t.TempDir(), d.p)
			checksums := checksum.New()
			for _, v := range d.m {
				checksums.Set(v.ID, v)
			}
			if err := checksums.UpdateFile(p); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			// The checksums must be readable again, so that the file aqua wrote
			// is the file aqua reads on the next run.
			read := checksum.New()
			if err := read.ReadFile(p); err != nil {
				t.Fatal(err)
			}
			for _, v := range d.m {
				if diff := cmp.Diff(v, read.Get(v.ID)); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestGetChecksumFilePathFromConfigFilePath(t *testing.T) {
	t.Parallel()
	// The paths are relative to a directory created for each test case.
	data := []struct {
		name        string
		cfgFilePath string
		exp         string
		files       map[string]string
	}{
		{
			name:        "new",
			cfgFilePath: fileAquaYaml,
			exp:         fileAquaChecksums,
		},
		{
			name:        "aqua-checksums.json > .aqua-checksums.json",
			cfgFilePath: fileAquaYaml,
			exp:         fileAquaChecksums,
			files: map[string]string{
				fileAquaChecksums:    "",
				fileDotAquaChecksums: "",
			},
		},
		{
			name:        fileDotAquaChecksums,
			cfgFilePath: fileAquaYaml,
			exp:         fileDotAquaChecksums,
			files: map[string]string{
				fileDotAquaChecksums: "",
			},
		},
		{
			name:        "a configuration file in a sub directory",
			cfgFilePath: "foo/" + fileAquaYaml,
			exp:         "foo/" + fileDotAquaChecksums,
			files: map[string]string{
				"foo/" + fileDotAquaChecksums: "",
				fileAquaChecksums:             "",
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			testutil.WriteFiles(t, dir, d.files)
			p, err := checksum.GetChecksumFilePathFromConfigFilePath(filepath.Join(dir, filepath.FromSlash(d.cfgFilePath)))
			if err != nil {
				t.Fatal(err)
			}
			if exp := filepath.Join(dir, filepath.FromSlash(d.exp)); p != exp {
				t.Fatalf("wanted %s, got %s", exp, p)
			}
		})
	}
}
