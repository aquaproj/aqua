package registry

type Cosign struct {
	CosignExperimental bool     `yaml:"cosign_experimental" json:"cosign_experimental,omitempty"`
	Opts               []string `json:"opts"`
}

func (cos *Cosign) GetEnabled() bool {
	if cos == nil {
		return false
	}
	return len(cos.Opts) != 0
}
