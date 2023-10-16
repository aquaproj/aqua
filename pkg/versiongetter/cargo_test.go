package versiongetter_test

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter"
	"github.com/google/go-cmp/cmp"
)

func TestCargoVersionGetter_Get(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		versions map[string][]string
		pkg      *registry.PackageInfo
		filters  []*versiongetter.Filter
		isErr    bool
		version  string
	}{
		{
			name: "normal",
			filters: []*versiongetter.Filter{
				{},
			},
			versions: map[string][]string{
				"crates.io/skim": {
					"3.0.0",
					"2.0.0",
					"1.0.0",
				},
			},
			pkg: &registry.PackageInfo{
				Crate: "crates.io/skim",
			},
			version: "3.0.0",
		},
	}

	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			cargoClient := versiongetter.NewMockCargoClient(d.versions)
			cargoGetter := versiongetter.NewCargo(cargoClient)
			version, err := cargoGetter.Get(ctx, d.pkg, d.filters)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if version != d.version {
				t.Fatalf("wanted %s, got %s", d.version, version)
			}
		})
	}
}

func TestCargoVersionGetter_List(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		versions map[string][]string
		pkg      *registry.PackageInfo
		filters  []*versiongetter.Filter
		isErr    bool
		items    []*fuzzyfinder.Item
	}{
		{
			name: "normal",
			filters: []*versiongetter.Filter{
				{},
			},
			versions: map[string][]string{
				"crates.io/skim": {
					"3.0.0",
					"2.0.0",
					"1.0.0",
				},
			},
			pkg: &registry.PackageInfo{
				Crate: "crates.io/skim",
			},
			items: []*fuzzyfinder.Item{
				{
					Item: "3.0.0",
				},
				{
					Item: "2.0.0",
				},
				{
					Item: "1.0.0",
				},
			},
		},
	}

	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			cargoClient := versiongetter.NewMockCargoClient(d.versions)
			cargoGetter := versiongetter.NewCargo(cargoClient)
			items, err := cargoGetter.List(ctx, d.pkg, d.filters)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if diff := cmp.Diff(items, d.items); diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}
