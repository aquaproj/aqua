package installpackage_test

const (
	errFileAlreadyExists = "file already exists"

	osLinux   = "linux"
	archAmd64 = "amd64"

	pathRoot       = "/home/foo/.local/share/aquaproj-aqua"
	pathRootBin    = "/home/foo/.local/share/aquaproj-aqua/bin"
	pathWorkspace  = "/home/foo/workspace"
	pathHomeFooFoo = "/home/foo/foo"
	fileAquaYaml   = "aqua.yaml"
	pkgFoo         = "foo"

	regTypeStandard      = "standard"
	pkgTypeGitHubContent = "github_content"
	pkgTypeGitHubRelease = "github_release"

	regOwnerAquaproj    = "aquaproj"
	regNameAquaRegistry = "aqua-registry"
	regFileRegistryYaml = "registry.yaml"

	repoSuzukiShunsuke       = "suzuki-shunsuke"
	repoSuzukiShunsukeCiInfo = "suzuki-shunsuke/ci-info"
	repoNameCiInfo           = "ci-info"
	versionV215              = "v2.15.0"
	versionV203              = "v2.0.3"
	tmplCiInfoAsset          = "ci-info_{{trimV .Version}}_{{.OS}}_amd64.tar.gz"
	pathCiInfoBinary         = "/home/foo/.local/share/aquaproj-aqua/pkgs/github_release/github.com/suzuki-shunsuke/ci-info/v2.0.3/ci-info_2.0.3_linux_amd64.tar.gz/ci-info"
)
