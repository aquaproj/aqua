package config

import (
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

// RenderCosign renders Cosign configuration with template variables.
// It processes Cosign options through templates to generate platform-specific signing configurations.
func (p *Package) RenderCosign(cos *registry.Cosign, rt *runtime.Runtime) (*registry.Cosign, error) {
	if p == nil || p.PackageInfo == nil || !cos.GetEnabled() {
		return nil, nil //nolint:nilnil
	}

	opts := make([]string, len(cos.Opts))
	for i, opt := range cos.Opts {
		s, err := p.RenderTemplateString(opt, rt)
		if err != nil {
			return nil, err
		}
		opts[i] = s
	}

	return &registry.Cosign{
		Signature:   cos.Signature,
		Certificate: cos.Certificate,
		Key:         cos.Key,
		Opts:        opts,
	}, nil
}
