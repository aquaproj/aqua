package checksum

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

func GetRuntimesFromSupportedEnvs(cfgSupportedEnvs, pkgSupportedEnvs []string) ([]*runtime.Runtime, error) {
	rts, err := runtime.GetRuntimesFromEnvs(pkgSupportedEnvs)
	if err != nil {
		return nil, fmt.Errorf("get supported platforms: %w", err)
	}
	if len(cfgSupportedEnvs) == 0 {
		return rts, nil
	}
	cfgRTs, err := runtime.GetRuntimesFromEnvs(cfgSupportedEnvs)
	if err != nil {
		return nil, fmt.Errorf("get supported platforms: %w", err)
	}

	cfgRTMap := make(map[string]struct{}, len(cfgRTs))
	for _, rt := range cfgRTs {
		cfgRTMap[rt.Env()] = struct{}{}
	}

	rtMap := make(map[string]*runtime.Runtime, len(rts))
	for _, rt := range rts {
		env := rt.Env()
		if _, ok := cfgRTMap[env]; ok {
			rtMap[env] = rt
		}
	}

	rts = make([]*runtime.Runtime, 0, len(rtMap))
	for _, rt := range rtMap {
		rts = append(rts, rt)
	}
	return rts, nil
}
