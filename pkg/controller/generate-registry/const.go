package genrgst

const (
	pkgTypeCargo         = "cargo"
	pkgTypeGitHubRelease = "github_release"

	flagCertOIDCIssuer     = "--certificate-oidc-issuer"
	flagCertIdentityRegexp = "--certificate-identity-regexp"
	flagSignature          = "--signature"
	urlOIDCIssuer          = "https://token.actions.githubusercontent.com"
	fileCosignPub          = "cosign.pub"

	// assetStateUploaded is the state of a GitHub Release asset whose upload has completed.
	// Assets in any other state (e.g. "starter") are invisible in the release page and aren't downloadable.
	assetStateUploaded = "uploaded"
)
