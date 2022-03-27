package validate

import (
	"errors"
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
)

func TestConfig(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		cfg   *config.Config
	}{
		{
			title: "normal",
			cfg: &config.Config{
				Packages: []*config.Package{
					{
						Name:     "suzuki-shunsuke/ci-info",
						Registry: "standard",
						Version:  "v1.0.0",
					},
				},
				Registries: config.Registries{
					"standard": &config.Registry{
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
		registries config.Registries
		isErr      bool
	}{
		{
			title: "normal",
			registries: config.Registries{
				"ci-info": &config.Registry{
					Name:      "ci-info",
					RepoOwner: "suzuki-shunsuke",
					RepoName:  "ci-info",
					Ref:       "v1.0.0",
					Path:      "registry.yaml",
					Type:      "github_content",
				},
				"standard": &config.Registry{
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
		pkgs  []*config.Package
		isErr bool
	}{
		{
			title: "normal",
			pkgs: []*config.Package{
				{
					Name:     "suzuki-shunsuke/cmdx",
					Registry: "standard",
				},
			},
		},
		{
			title: "duplicated",
			pkgs: []*config.Package{
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
		pkgInfos config.PackageInfos
		isErr    bool
	}{
		{
			title: "normal",
			pkgInfos: config.PackageInfos{
				&config.PackageInfo{
					Name: "foo",
					Files: []*config.File{
						{
							Name: "foo",
						},
					},
				},
				&config.PackageInfo{
					Name: "bar",
					Files: []*config.File{
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
			pkgInfos: config.PackageInfos{
				&config.PackageInfo{
					Name: "foo",
					Files: []*config.File{
						{
							Name: "foo",
						},
					},
				},
				&config.PackageInfo{
					Name: "foo",
					Files: []*config.File{
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
