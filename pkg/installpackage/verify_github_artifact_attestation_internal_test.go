package installpackage

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

// The attestation is looked up in the repository the configuration names, or the package's.
func TestGitHubArtifactAttestationsVerifier_repository(t *testing.T) {
	t.Parallel()
	pkg := &config.Package{PackageInfo: &registry.PackageInfo{RepoOwner: "cli", RepoName: "cli"}}
	tests := []struct {
		name string
		gaa  *registry.GitHubArtifactAttestations
		want string
	}{
		{name: "default", gaa: &registry.GitHubArtifactAttestations{}, want: "cli/cli"},
		{name: "set", gaa: &registry.GitHubArtifactAttestations{Repository: "org/other"}, want: "org/other"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := &gitHubArtifactAttestationsVerifier{gaa: tt.gaa, pkg: pkg}
			if got := g.repository(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
