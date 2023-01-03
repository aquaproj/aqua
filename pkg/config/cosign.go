package config

import (
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/runtime"
)

func (cpkg *Package) RenderCosign(cos *registry.Cosign, rt *runtime.Runtime) (*registry.Cosign, error) {
	if cpkg == nil || cpkg.PackageInfo == nil || !cos.GetEnabled() {
		return nil, nil //nolint:nilnil
	}

	opts := make([]string, len(cos.Opts))
	for i, opt := range cos.Opts {
		s, err := cpkg.RenderTemplateString(opt, rt)
		if err != nil {
			return nil, err
		}
		opts[i] = s
	}

	return &registry.Cosign{
		CosignExperimental: cos.CosignExperimental,
		Signature:          cos.Signature,
		Certificate:        cos.Certificate,
		Key:                cos.Key,
		Opts:               opts,
	}, nil
}
