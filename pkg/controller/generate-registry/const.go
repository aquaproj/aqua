package genrgst

const (
	pkgTypeCargo         = "cargo"
	pkgTypeGitHubRelease = "github_release"

	flagCertOIDCIssuer     = "--certificate-oidc-issuer"
	flagCertIdentityRegexp = "--certificate-identity-regexp"
	flagSignature          = "--signature"
	urlOIDCIssuer          = "https://token.actions.githubusercontent.com"
	fileCosignPub          = "cosign.pub"
)
