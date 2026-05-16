package update_test

const (
	pkgNameTfcmt      = "tfcmt"
	repoSuzukiTfcmt   = "suzuki-shunsuke/tfcmt"
	repoCliCli        = "cli/cli"
	pathWorkspaceYaml = "/workspace/aqua.yaml"
	pathWorkspace     = "/workspace"

	regTypeStandard      = "standard"
	pkgTypeGitHubRelease = "github_release"
	pkgTypeGitHubContent = "github_content"

	tmplTfcmtAsset = "tfcmt_{{.OS}}_{{.Arch}}.tar.gz"
	tmplGhAsset    = "gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.zip"
	tmplGhBinSrc   = "{{.AssetWithoutExt}}/bin/gh"

	repoOwnerSuzuki     = "suzuki-shunsuke"
	repoOwnerAquaproj   = "aquaproj"
	regNameAquaRegistry = "aqua-registry"
	regFileRegistryYaml = "registry.yaml"
	repoOwnerCli        = "cli"

	osDarwin  = "darwin"
	archArm64 = "arm64"
)
