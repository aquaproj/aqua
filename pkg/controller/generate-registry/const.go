package genrgst

import "github.com/aquaproj/aqua/v2/pkg/config/registry"

// The types a generated definition can say, named where a definition is read so that one
// of them changing is one place changing.
const (
	pkgTypeCargo         = registry.PkgInfoTypeCargo
	pkgTypeGitHubRelease = registry.PkgInfoTypeGitHubRelease
	pkgTypeGitLabRelease = registry.PkgInfoTypeGitLabRelease

	flagCertOIDCIssuer     = "--certificate-oidc-issuer"
	flagCertIdentityRegexp = "--certificate-identity-regexp"
	flagSignature          = "--signature"
	urlOIDCIssuer          = "https://token.actions.githubusercontent.com"
	fileCosignPub          = "cosign.pub"

	// assetStateUploaded is the state of a GitHub Release asset whose upload has completed.
	// Assets in any other state (e.g. "starter") are invisible in the release page and aren't downloadable.
	assetStateUploaded = "uploaded"
)
