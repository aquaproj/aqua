package registry

// BundleAssetPackslip is the name packslip publishes its signed bundle under, in the
// release whose artifacts it signs.
const BundleAssetPackslip = "packslip.sigstore.json"

// Packslip defines configuration for verifying packages with packslip.
// packslip publishes a signed release bundle, packslip.sigstore.json, beside the
// release's artifacts, and aqua downloads it from the package's own release
// unless URL says where it is.
// https://github.com/jdx/packslip
type Packslip struct {
	// Enabled controls whether packslip verification is active.
	Enabled *bool `yaml:",omitempty" json:"enabled,omitempty"`
	// URL is the URL of the packslip bundle.
	// If it isn't set, packslip.sigstore.json is downloaded from the package's release.
	// It is required for a package that isn't a release asset, such as the http type.
	URL *string `yaml:",omitempty" json:"url,omitempty"`
	// PublicKey is the public key of a key-signed release, as its base64 line.
	// It is passed to `packslip verify --pubkey`.
	PublicKey string `yaml:"pubkey,omitempty" json:"pubkey,omitempty"`
	// Identity is the exact certificate identity a keyless signer must have, including its ref.
	// It is passed to `packslip verify --identity`.
	Identity string `yaml:",omitempty" json:"identity,omitempty"`
	// IdentityPrefix is a prefix the certificate identity must start with, such as https://github.com/owner/repo/.
	// It is passed to `packslip verify --identity-prefix`.
	IdentityPrefix string `yaml:"identity_prefix,omitempty" json:"identity_prefix,omitempty"`
	// Issuer is the OIDC issuer a keyless signer must have, such as https://token.actions.githubusercontent.com.
	// It is passed to `packslip verify --issuer`.
	Issuer string `yaml:",omitempty" json:"issuer,omitempty"`
	// Pin is the signer fingerprint (ps1_...) the certificate's repository must have.
	// It is passed to `packslip verify --pin`.
	Pin string `yaml:",omitempty" json:"pin,omitempty"`
	// AllowUnlogged accepts a bundle without a transparency log entry.
	// It is passed to `packslip verify --allow-unlogged`.
	AllowUnlogged bool `yaml:"allow_unlogged,omitempty" json:"allow_unlogged,omitempty"`
}

// GetEnabled returns whether packslip verification is enabled.
// If Enabled is nil, it defaults to true.
func (p *Packslip) GetEnabled() bool {
	if p == nil {
		return false
	}
	if p.Enabled != nil {
		return *p.Enabled
	}
	return true
}

// ToDownloadedFile converts the packslip configuration to a DownloadedFile,
// which is used for downloading the bundle.
//
// pkgType is the package's type. Without URL, the bundle is the release's own
// packslip.sigstore.json, in the same release as the package's asset.
func (p *Packslip) ToDownloadedFile(pkgType string) *DownloadedFile {
	if p.URL != nil {
		return &DownloadedFile{
			Type: PkgInfoTypeHTTP,
			URL:  p.URL,
		}
	}
	asset := BundleAssetPackslip
	return &DownloadedFile{
		Type:  pkgType,
		Asset: &asset,
	}
}

// HasTrust reports whether the configuration says whom to trust.
//
// Without a key, an identity, or a fingerprint, packslip derives its policy from the
// project the bundle itself claims, which proves only that the signer belongs to that
// project, not that it is the package aqua was asked to install.
func (p *Packslip) HasTrust() bool {
	return p.PublicKey != "" || p.Identity != "" || p.IdentityPrefix != "" || p.Pin != ""
}
