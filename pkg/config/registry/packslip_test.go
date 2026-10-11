package registry_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

func TestPackageInfo_Validate_packslip(t *testing.T) { //nolint:funlen
	t.Parallel()
	release := func(ps *registry.Packslip) *registry.PackageInfo {
		return &registry.PackageInfo{
			Type:      registry.PkgInfoTypeGitHubRelease,
			RepoOwner: "jdx",
			RepoName:  "hk",
			Asset:     "hk-{{.Arch}}-{{.OS}}.{{.Format}}",
			Packslip:  ps,
		}
	}
	http := func(ps *registry.Packslip) *registry.PackageInfo {
		return &registry.PackageInfo{
			Type:      registry.PkgInfoTypeHTTP,
			RepoOwner: "jdx",
			RepoName:  "hk",
			URL:       "https://example.com/hk-{{.Version}}.tar.gz",
			Packslip:  ps,
		}
	}
	data := []struct {
		title   string
		pkgInfo *registry.PackageInfo
		isErr   bool
	}{
		{
			title: "identity_prefix",
			pkgInfo: release(&registry.Packslip{
				IdentityPrefix: "https://github.com/jdx/hk/",
				Issuer:         "https://token.actions.githubusercontent.com",
			}),
		},
		{
			// packslip would otherwise trust whichever project the bundle claims.
			title: "no trust",
			pkgInfo: release(&registry.Packslip{
				Issuer: "https://token.actions.githubusercontent.com",
			}),
			isErr: true,
		},
		{
			title: "disabled",
			pkgInfo: release(&registry.Packslip{
				Enabled: new(false),
			}),
		},
		{
			title: "http with url",
			pkgInfo: http(&registry.Packslip{
				URL: new("https://example.com/packslip.sigstore.json"),
				Pin: "ps1_snirenkjwr7m5ozgcufameodnm",
			}),
		},
		{
			// An http package has no release to find the bundle in.
			title: "http without url",
			pkgInfo: http(&registry.Packslip{
				Pin: "ps1_snirenkjwr7m5ozgcufameodnm",
			}),
			isErr: true,
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if err := d.pkgInfo.Validate(); err != nil {
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
