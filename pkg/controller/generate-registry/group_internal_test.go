package genrgst

import (
	"reflect"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/hashicorp/go-version"
)

func Test_sortAndMergeGroups(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name   string
		groups []*Group
		exp    []*Group
	}{
		{
			name: "normal",
			groups: []*Group{
				{
					releases: []*Release{
						{
							Tag:     "v0.1.0",
							Version: version.Must(version.NewSemver("v0.1.0")),
						},
						{
							Tag:     "v0.2.0",
							Version: version.Must(version.NewSemver("v0.2.0")),
						},
					},
					allAsset: `tfcmt_{{.Version}}_darwin_amd64.tar.gz
tfcmt_{{.Version}}_linux_amd64.tar.gz
`,
					pkg: &Package{
						Info: &registry.PackageInfo{
							Asset: "tfcmt_{{.Version}}_{{.OS}}_{{.Arch}}.tar.gz",
						},
					},
				},
				{
					releases: []*Release{
						{
							Tag:     "v0.3.0",
							Version: version.Must(version.NewSemver("v0.3.0")),
						},
					},
					allAsset: `tfcmt_darwin_amd64.tar.gz
tfcmt_linux_amd64.tar.gz
`,
					pkg: &Package{
						Info: &registry.PackageInfo{
							Asset: "tfcmt_{{.OS}}_{{.Arch}}.tar.gz",
						},
					},
				},
				{
					releases: []*Release{
						{
							Tag:     "v0.4.0",
							Version: version.Must(version.NewSemver("v0.4.0")),
						},
						{
							Tag:     "v0.5.0",
							Version: version.Must(version.NewSemver("v0.5.0")),
						},
					},
					allAsset: `tfcmt_{{.Version}}_darwin_amd64.tar.gz
tfcmt_{{.Version}}_linux_amd64.tar.gz
`,
					pkg: &Package{
						Info: &registry.PackageInfo{
							Asset: "tfcmt_{{.Version}}_{{.OS}}_{{.Arch}}.tar.gz",
						},
					},
				},
				{
					releases: []*Release{
						{
							Tag:     "v0.6.0",
							Version: version.Must(version.NewSemver("v0.6.0")),
						},
					},
					allAsset: `tfcmt_darwin_amd64.tar.gz
tfcmt_linux_amd64.tar.gz
tfcmt_windows_amd64.tar.gz
`,
					pkg: &Package{
						Info: &registry.PackageInfo{
							Asset: "tfcmt_{{.OS}}_{{.Arch}}.tar.gz",
						},
					},
				},
			},
			exp: []*Group{
				{
					releases: []*Release{
						{
							Tag:     "v0.3.0",
							Version: version.Must(version.NewSemver("v0.3.0")),
						},
					},
					allAsset: `tfcmt_darwin_amd64.tar.gz
tfcmt_linux_amd64.tar.gz
`,
					fixed: true,
					pkg: &Package{
						Info: &registry.PackageInfo{
							Asset: "tfcmt_{{.OS}}_{{.Arch}}.tar.gz",
						},
					},
				},
				{
					releases: []*Release{
						{
							Tag:     "v0.1.0",
							Version: version.Must(version.NewSemver("v0.1.0")),
						},
						{
							Tag:     "v0.2.0",
							Version: version.Must(version.NewSemver("v0.2.0")),
						},
						{
							Tag:     "v0.4.0",
							Version: version.Must(version.NewSemver("v0.4.0")),
						},
						{
							Tag:     "v0.5.0",
							Version: version.Must(version.NewSemver("v0.5.0")),
						},
					},
					allAsset: `tfcmt_{{.Version}}_darwin_amd64.tar.gz
tfcmt_{{.Version}}_linux_amd64.tar.gz
`,
					pkg: &Package{
						Info: &registry.PackageInfo{
							Asset: "tfcmt_{{.Version}}_{{.OS}}_{{.Arch}}.tar.gz",
						},
					},
				},
				{
					releases: []*Release{
						{
							Tag:     "v0.6.0",
							Version: version.Must(version.NewSemver("v0.6.0")),
						},
					},
					allAsset: `tfcmt_darwin_amd64.tar.gz
tfcmt_linux_amd64.tar.gz
tfcmt_windows_amd64.tar.gz
`,
					pkg: &Package{
						Info: &registry.PackageInfo{
							Asset: "tfcmt_{{.OS}}_{{.Arch}}.tar.gz",
						},
					},
				},
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			groups := sortAndMergeGroups(d.groups)
			if !reflect.DeepEqual(d.exp, groups) {
				t.Errorf("groups are unexpected")
			}
		})
	}
}

func Test_groupByExcludedAsset(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name   string
		groups []*Group
		exp    []*Group
	}{
		{
			name: "normal",
			groups: []*Group{
				{
					releases: []*Release{
						{
							Tag: "v0.1.0",
						},
					},
					allAsset: `tfcmt_{{.Version}}_darwin_amd64.tar.gz
tfcmt_{{.Version}}_linux_amd64.tar.gz
`,
				},
				{
					releases: []*Release{
						{
							Tag: "v0.2.0",
						},
					},
					allAsset: `tfcmt_darwin_amd64.tar.gz
tfcmt_linux_amd64.tar.gz
`,
				},
				{
					releases: []*Release{
						{
							Tag: "v0.3.0",
						},
					},
					allAsset: `tfcmt_darwin_amd64.tar.gz
tfcmt_linux_amd64.tar.gz
`,
				},
				{
					releases: []*Release{
						{
							Tag: "v0.4.0",
						},
					},
					allAsset: `tfcmt_darwin_amd64.tar.gz
tfcmt_linux_amd64.tar.gz
tfcmt_windows_amd64.tar.gz
`,
				},
			},
			exp: []*Group{
				{
					releases: []*Release{
						{
							Tag: "v0.1.0",
						},
					},
					allAsset: `tfcmt_{{.Version}}_darwin_amd64.tar.gz
tfcmt_{{.Version}}_linux_amd64.tar.gz
`,
				},
				{
					releases: []*Release{
						{
							Tag: "v0.2.0",
						},
						{
							Tag: "v0.3.0",
						},
					},
					allAsset: `tfcmt_darwin_amd64.tar.gz
tfcmt_linux_amd64.tar.gz
`,
				},
				{
					releases: []*Release{
						{
							Tag: "v0.4.0",
						},
					},
					allAsset: `tfcmt_darwin_amd64.tar.gz
tfcmt_linux_amd64.tar.gz
tfcmt_windows_amd64.tar.gz
`,
				},
			},
		},
		{
			name: "normal 2",
			groups: []*Group{
				{
					releases: []*Release{
						{
							Tag: "v0.1.0",
						},
						{
							Tag: "v0.2.0",
						},
					},
					allAsset: `tfcmt_{{.Version}}_darwin_amd64.tar.gz
tfcmt_{{.Version}}_linux_amd64.tar.gz
`,
				},
				{
					releases: []*Release{
						{
							Tag: "v0.4.0",
						},
						{
							Tag: "v0.5.0",
						},
					},
					allAsset: `tfcmt_{{.Version}}_darwin_amd64.tar.gz
tfcmt_{{.Version}}_linux_amd64.tar.gz
`,
				},
			},
			exp: []*Group{
				{
					releases: []*Release{
						{
							Tag: "v0.1.0",
						},
						{
							Tag: "v0.2.0",
						},
						{
							Tag: "v0.4.0",
						},
						{
							Tag: "v0.5.0",
						},
					},
					allAsset: `tfcmt_{{.Version}}_darwin_amd64.tar.gz
tfcmt_{{.Version}}_linux_amd64.tar.gz
`,
				},
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			groups := groupByExcludedAsset(d.groups)
			if !reflect.DeepEqual(d.exp, groups) {
				t.Errorf("groups are unexpected")
			}
		})
	}
}
