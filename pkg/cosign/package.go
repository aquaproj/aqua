package cosign

import (
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

func Package() *config.Package {
	return &config.Package{
		Package: &aqua.Package{
			Name:    "sigstore/cosign",
			Version: Version,
		},
		PackageInfo: &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: "sigstore",
			RepoName:  "cosign",
			Asset:     "cosign-{{.OS}}-{{.Arch}}",
			SupportedEnvs: []string{
				"darwin",
				"linux",
				"amd64",
			},
		},
	}
}
