package validate

import (
	"errors"
	"testing"

	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
)

func TestConfig(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		cfg   *aqua.Config
	}{
		{
			title: "normal",
			cfg: &aqua.Config{
				Packages: []*aqua.Package{
					{
						Name:     "suzuki-shunsuke/ci-info",
						Registry: "standard",
						Version:  "v1.0.0",
					},
				},
				Registries: aqua.Registries{
					"standard": &aqua.Registry{
						Name:      "standard",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Ref:       "v0.8.0",
						Path:      "registry.yaml",
						Type:      "github_content",
					},
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if err := Config(d.cfg); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_validateRegistries(t *testing.T) {
	t.Parallel()
	data := []struct {
		title      string
		registries aqua.Registries
		isErr      bool
	}{
		{
			title: "normal",
			registries: aqua.Registries{
				"ci-info": &aqua.Registry{
					Name:      "ci-info",
					RepoOwner: "suzuki-shunsuke",
					RepoName:  "ci-info",
					Ref:       "v1.0.0",
					Path:      "registry.yaml",
					Type:      "github_content",
				},
				"standard": &aqua.Registry{
					Name:      "standard",
					RepoOwner: "aquaproj",
					RepoName:  "aqua-registry",
					Ref:       "v0.8.0",
					Path:      "registry.yaml",
					Type:      "github_content",
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			err := validateRegistries(d.registries)
			if d.isErr {
				if !errors.Is(err, ErrRegistryNameIsDuplicated) {
					t.Fatal(err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_validatePackages(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		pkgs  []*aqua.Package
		isErr bool
	}{
		{
			title: "normal",
			pkgs: []*aqua.Package{
				{
					Name:     "suzuki-shunsuke/cmdx",
					Registry: "standard",
				},
			},
		},
		{
			title: "duplicated",
			pkgs: []*aqua.Package{
				{
					Name:     "suzuki-shunsuke/cmdx",
					Registry: "standard",
				},
				{
					Name:     "suzuki-shunsuke/cmdx",
					Registry: "standard",
				},
			},
			isErr: false,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			err := validatePackages(d.pkgs)
			if d.isErr {
				t.Fatal(err)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPackageInfos(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title    string
		pkgInfos registry.PackageInfos
		isErr    bool
	}{
		{
			title: "normal",
			pkgInfos: registry.PackageInfos{
				&registry.PackageInfo{
					Name: "foo",
					Files: []*registry.File{
						{
							Name: "foo",
						},
					},
				},
				&registry.PackageInfo{
					Name: "bar",
					Files: []*registry.File{
						{
							Name: "bar",
						},
					},
				},
			},
		},
		{
			title: "duplicated",
			isErr: true,
			pkgInfos: registry.PackageInfos{
				&registry.PackageInfo{
					Name: "foo",
					Files: []*registry.File{
						{
							Name: "foo",
						},
					},
				},
				&registry.PackageInfo{
					Name: "foo",
					Files: []*registry.File{
						{
							Name: "foo",
						},
					},
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			err := validatePackageInfos(d.pkgInfos)
			if d.isErr {
				if !errors.Is(err, errPkgNameMustBeUniqueInRegistry) {
					t.Fatal(err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
