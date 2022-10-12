package security

type Config struct {
	// OR condition
	Registries []*Registry `json:"registries,omitempty"`
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
	ID string `json:"id"`
	// AND condition
	IDFormat string `yaml:"id_format" json:"id_format,omitempty"`
	Version  string `json:"version,omitempty"`
}
