package policy

type Config struct {
	// OR condition
	Packages []*Package `json:"packages,omitempty"`
}

type Registry struct {
	ID string `json:"id"`
	// AND condition
	IDFormat string `yaml:"id_format" json:"id_format,omitempty"`
	Version  string `json:"version,omitempty"`
}

type Package struct {
	Packages []*ChildPackage `json:"packages,omitempty"`
	Registry *Registry       `json:"registry,omitempty"`
}

type ChildPackage struct {
	ID string `json:"id"`
	// AND condition
	IDFormat string `yaml:"id_format" json:"id_format,omitempty"`
	Version  string `json:"version,omitempty"`
}
