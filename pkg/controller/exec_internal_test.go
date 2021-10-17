package controller

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestController_findExecFileFromPkg(t *testing.T) {
	t.Parallel()
	data := []struct {
		title          string
		registries     map[string]*RegistryContent
		exeName        string
		pkg            *Package
		expPackageInfo *MergedPackageInfo
		expFile        *File
	}{
		{
			title:   "normal",
			exeName: "kubectl",
			pkg: &Package{
				Registry: "standard",
				Name:     "kubernetes/kubectl",
			},
			expPackageInfo: &MergedPackageInfo{
				Name: "kubernetes/kubectl",
				Files: []*File{
					{
						Name: "kubectl",
					},
				},
			},
			expFile: &File{
				Name: "kubectl",
			},
			registries: map[string]*RegistryContent{
				"standard": {
					PackageInfos: PackageInfos{
						&MergedPackageInfo{
							Name: "kubernetes/kubectl",
							Files: []*File{
								{
									Name: "kubectl",
								},
							},
						},
					},
				},
			},
		},
	}
	ctrl := &Controller{}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkgInfo, file := ctrl.findExecFileFromPkg(d.registries, d.exeName, d.pkg)
			if diff := cmp.Diff(d.expPackageInfo, pkgInfo); diff != "" {
				t.Fatal(diff)
			}
			if diff := cmp.Diff(d.expFile, file); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestController_execCommand(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exePath string
		args    []string
	}{
		{
			title:   "normal",
			exePath: "echo",
			args:    []string{"hello"},
		},
	}
	ctrl := &Controller{}
	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			err := ctrl.execCommand(ctx, d.exePath, d.args)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
