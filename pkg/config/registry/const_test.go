package registry_test

const (
	caseDefault      = "default"
	caseEnabledTrue  = "enabled explicitly true"
	caseEnabledFalse = "enabled explicitly false"

	osLinuxCapitalized = "Linux"
	osDarwin           = "darwin"
	archAmd64          = "amd64"
	archArm64          = "arm64"

	pkgTypeGitHubRelease = "github_release"
	pkgTypeHTTP          = "http"

	pkgFoo            = "foo"
	pkgNameAqua       = "aqua"
	repoOwnerAquaproj = "aquaproj"
	repoNameCiInfo    = "ci-info"

	flagRekorURL     = "--rekor-url"
	flagCertIdentity = "--certificate-identity"

	urlSlsaDev        = "https://slsa.dev/provenance/v0.2"
	pathReleaseYml    = ".github/workflows/release.yml@refs/tags/{{.Version}}"
	pathDeprecatedYml = ".github/workflows/deprecated.yml@refs/tags/v1.0.0"
)
