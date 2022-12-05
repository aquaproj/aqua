package registry

type Cosign struct {
	Opts []string `json:"opts"`
}

func (cos *Cosign) GetEnabled() bool {
	if cos == nil {
		return false
	}
	return len(cos.Opts) != 0
}
