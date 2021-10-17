package controller

import (
	"errors"
	"testing"
)

func Test_validateConfig(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		cfg   *Config
	}{
		{
			title: "normal",
			cfg: &Config{
				Packages: []*Package{
					{
						Name:     "suzuki-shunsuke/ci-info",
						Registry: "standard",
						Version:  "v1.0.0",
					},
				},
				Registries: Registries{
					&GitHubContentRegistry{
						Name:      "standard",
						RepoOwner: "suzuki-shunsuke",
						RepoName:  "aqua-registry",
						Ref:       "v0.8.0",
						Path:      "registry.yaml",
					},
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			err := validateConfig(d.cfg)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_validateRegistries(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title      string
		registries Registries
		isErr      bool
	}{
		{
			title: "normal",
			registries: Registries{
				&GitHubContentRegistry{
					Name:      "suzuki-shunsuke/ci-info",
					RepoOwner: "suzuki-shunsuke",
					RepoName:  "ci-info",
					Ref:       "v1.0.0",
					Path:      "registry.yaml",
				},
				&GitHubContentRegistry{
					Name:      "suzuki-shunsuke/aqua-registry",
					RepoOwner: "suzuki-shunsuke",
					RepoName:  "aqua-registry",
					Ref:       "v0.8.0",
					Path:      "registry.yaml",
				},
			},
		},
		{
			title: "duplicated",
			isErr: true,
			registries: Registries{
				&GitHubContentRegistry{
					Name:      "suzuki-shunsuke/ci-info",
					RepoOwner: "suzuki-shunsuke",
					RepoName:  "ci-info",
					Ref:       "v1.0.0",
					Path:      "registry.yaml",
				},
				&GitHubContentRegistry{
					Name:      "suzuki-shunsuke/ci-info",
					RepoOwner: "suzuki-shunsuke",
					RepoName:  "ci-info",
					Ref:       "v0.8.0",
					Path:      "registry.yaml",
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
				if !errors.Is(err, errRegistryNameIsDuplicated) {
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
		pkgs  []*Package
		isErr bool
	}{
		{
			title: "normal",
			pkgs: []*Package{
				{
					Name:     "suzuki-shunsuke/cmdx",
					Registry: "standard",
				},
			},
		},
		{
			title: "duplicated",
			pkgs: []*Package{
				{
					Name:     "suzuki-shunsuke/cmdx",
					Registry: "standard",
				},
				{
					Name:     "suzuki-shunsuke/cmdx",
					Registry: "standard",
				},
			},
			isErr: true,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			err := validatePackages(d.pkgs)
			if d.isErr {
				if !errors.Is(err, errPairPkgNameAndRegistryMustBeUnique) {
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

func Test_validatePackageInfos(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title    string
		pkgInfos PackageInfos
		isErr    bool
	}{
		{
			title: "normal",
			pkgInfos: PackageInfos{
				&MergedPackageInfo{
					Name: "foo",
					Files: []*File{
						{
							Name: "foo",
						},
					},
				},
				&MergedPackageInfo{
					Name: "bar",
					Files: []*File{
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
			pkgInfos: PackageInfos{
				&MergedPackageInfo{
					Name: "foo",
					Files: []*File{
						{
							Name: "foo",
						},
					},
				},
				&MergedPackageInfo{
					Name: "foo",
					Files: []*File{
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
