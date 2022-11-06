package policy

type Config struct {
	Registries []*Registry `json:"registries"`
	Packages   []*Package  `json:"packages,omitempty"`
}

type Registry struct {
	Type string `json:"type"`
	Ref  string `json:"ref"`
	Name string `json:"name,omitempty"`
}

type Package struct {
	Name         string    `json:"name"`
	Version      string    `json:"version,omitempty"`
	RegistryName string    `yaml:"registry" json:"registry,omitempty"`
	Registry     *Registry `yaml:"-" json:"-"`
}
