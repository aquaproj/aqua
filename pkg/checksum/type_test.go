package checksum_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
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
			fs, err := testutil.NewFs(d.m)
			if err != nil {
				t.Fatal(err)
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
			fs := afero.NewMemMapFs()
			checksums := checksum.New()
			for _, v := range d.m {
				checksums.Set(v.ID, v)
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
			name:        "new absolute",
			cfgFilePath: pathHomeFooAquaYaml,
			exp:         pathHomeFooAquaChecksums,
		},
		{
			name:        "absolute aqua-checksums.json > .aqua-checksums.json",
			cfgFilePath: pathHomeFooAquaYaml,
			exp:         pathHomeFooAquaChecksums,
			files: map[string]string{
				pathHomeFooDotAquaChecksums: "",
				pathHomeFooAquaChecksums:    "",
			},
		},
		{
			name:        "absolute .aqua-checksums.json",
			cfgFilePath: pathHomeFooAquaYaml,
			exp:         pathHomeFooDotAquaChecksums,
			files: map[string]string{
				pathHomeFooDotAquaChecksums: "",
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs, err := testutil.NewFs(d.files)
			if err != nil {
				t.Fatal(err)
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
