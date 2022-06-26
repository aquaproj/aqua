package registry

import (
	"github.com/clivm/clivm/pkg/expr"
	"github.com/clivm/clivm/pkg/runtime"
)

func (pkgInfo *PackageInfo) CheckSupported(rt *runtime.Runtime, env string) (bool, error) {
	if pkgInfo.SupportedEnvs != nil {
		return pkgInfo.CheckSupportedEnvs(rt.GOOS, rt.GOARCH, env), nil
	}
	if pkgInfo.SupportedIf == nil {
		return true, nil
	}
	return expr.EvaluateSupportedIf(pkgInfo.SupportedIf, rt) //nolint:wrapcheck
}

func (pkgInfo *PackageInfo) CheckSupportedEnvs(goos, goarch, env string) bool {
	if pkgInfo.SupportedEnvs == nil {
		return true
	}
	for _, supportedEnv := range pkgInfo.SupportedEnvs {
		switch supportedEnv {
		case goos, goarch, env, "all":
			return true
		}
	}
	return false
}
