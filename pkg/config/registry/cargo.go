package registry

// Cargo defines configuration options for installing Rust packages via Cargo.
// These options control how cargo install commands are executed.
type Cargo struct {
	// Features is a list of specific features to enable when installing.
	Features []string `yaml:",omitempty" json:"features,omitempty"`
	// AllFeatures enables all available features for the package.
	AllFeatures bool `yaml:"all_features" json:"all_features,omitempty"`
	// Locked uses the exact versions from Cargo.lock file.
	Locked bool `yaml:",omitempty" json:"locked,omitempty"`
}
