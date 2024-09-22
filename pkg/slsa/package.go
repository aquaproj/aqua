package slsa

import (
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

func Package() *config.Package {
	return &config.Package{
		Package: &aqua.Package{
			Name:    "slsa-framework/slsa-verifier",
			Version: Version,
		},
		PackageInfo: &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: "slsa-framework",
			RepoName:  "slsa-verifier",
			Asset:     "slsa-verifier-{{.OS}}-{{.Arch}}",
		},
	}
}
