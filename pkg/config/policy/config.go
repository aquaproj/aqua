package policy

type Config struct {
	Registries []*Registry `json:"registries"`
	Packages   []*Package  `json:"packages,omitempty"`
}

type Registry struct {
	Name      string `json:"name,omitempty"`
	Type      string `validate:"required" json:"type,omitempty" jsonschema:"enum=standard,enum=local,enum=github_content"`
	RepoOwner string `yaml:"repo_owner" json:"repo_owner,omitempty"`
	RepoName  string `yaml:"repo_name" json:"repo_name,omitempty"`
	Ref       string `json:"ref,omitempty"`
	Path      string `validate:"required" json:"path,omitempty"`
}

type Package struct {
	Name         string    `json:"name"`
	Version      string    `json:"version,omitempty"`
	RegistryName string    `yaml:"registry" json:"registry,omitempty"`
	Registry     *Registry `yaml:"-" json:"-"`
}
