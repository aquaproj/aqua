package registry

type Cargo struct {
	Features    []string `json:"features,omitempty"`
	AllFeatures bool     `yaml:"all_features" json:"all_features,omitempty"`
}
