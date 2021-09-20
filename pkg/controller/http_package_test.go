package controller_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/aqua/pkg/controller"
)

func TestHTTPPackageInfo_GetName(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *controller.HTTPPackageInfo
	}{
		{
			title: "normal",
			exp:   "foo",
			pkgInfo: &controller.HTTPPackageInfo{
				Name: "foo",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if name := d.pkgInfo.GetName(); name != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, name)
			}
		})
	}
}

func TestHTTPPackageInfo_GetType(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *controller.HTTPPackageInfo
	}{
		{
			title:   "normal",
			exp:     "http",
			pkgInfo: &controller.HTTPPackageInfo{},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if typ := d.pkgInfo.GetType(); typ != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, typ)
			}
		})
	}
}

func TestHTTPPackageInfo_GetLink(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *controller.HTTPPackageInfo
	}{
		{
			title: "normal",
			exp:   "http://example.com",
			pkgInfo: &controller.HTTPPackageInfo{
				Link: "http://example.com",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if link := d.pkgInfo.GetLink(); link != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, link)
			}
		})
	}
}

func TestHTTPPackageInfo_GetDescription(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *controller.HTTPPackageInfo
	}{
		{
			title: "normal",
			exp:   "hello world",
			pkgInfo: &controller.HTTPPackageInfo{
				Description: "hello world",
			},
		},
		{
			title:   "empty",
			exp:     "",
			pkgInfo: &controller.HTTPPackageInfo{},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if desc := d.pkgInfo.GetDescription(); desc != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, desc)
			}
		})
	}
}

func TestHTTPPackageInfo_GetFormat(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *controller.HTTPPackageInfo
	}{
		{
			title: "normal",
			exp:   "tar.gz",
			pkgInfo: &controller.HTTPPackageInfo{
				Format: "tar.gz",
			},
		},
		{
			title:   "empty",
			exp:     "",
			pkgInfo: &controller.HTTPPackageInfo{},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if format := d.pkgInfo.GetFormat(); format != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, format)
			}
		})
	}
}

func TestHTTPPackageInfo_GetReplacements(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     map[string]string
		pkgInfo *controller.HTTPPackageInfo
	}{
		{
			title: "normal",
			exp: map[string]string{
				"amd64": "x86_64",
			},
			pkgInfo: &controller.HTTPPackageInfo{
				Replacements: map[string]string{
					"amd64": "x86_64",
				},
			},
		},
		{
			title:   "empty",
			exp:     nil,
			pkgInfo: &controller.HTTPPackageInfo{},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			replacements := d.pkgInfo.GetReplacements()
			if diff := cmp.Diff(d.exp, replacements); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestHTTPPackageInfo_GetFiles(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     []*controller.File
		pkgInfo *controller.HTTPPackageInfo
	}{
		{
			title: "normal",
			exp: []*controller.File{
				{
					Name: "go",
				},
				{
					Name: "gofmt",
				},
			},
			pkgInfo: &controller.HTTPPackageInfo{
				Files: []*controller.File{
					{
						Name: "go",
					},
					{
						Name: "gofmt",
					},
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			files := d.pkgInfo.GetFiles()
			if diff := cmp.Diff(d.exp, files); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
