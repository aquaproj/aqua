package genrgst

import (
	"reflect"

	"github.com/aquaproj/aqua/pkg/config/registry"
)

func (ctrl *Controller) setSupportedEnvs(envs map[string]struct{}, pkgInfo *registry.PackageInfo) { //nolint:cyclop
	if has(envs, "darwin") || has(envs, "darwin/amd64") {
		pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, "darwin")
	} else if has(envs, "darwin/arm64") {
		pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, "darwin/arm64")
	}
	if has(envs, "darwin/amd64") && !has(envs, "darwin/arm64") {
		pkgInfo.Rosetta2 = boolP(true)
	}
	if has(envs, "linux/amd64") {
		if has(envs, "linux/arm64") {
			pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, "linux")
		} else {
			pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, "linux/amd64")
		}
	} else if has(envs, "linux/arm64") {
		pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, "linux/arm64")
	}
	if has(envs, "windows/amd64") {
		if has(envs, "windows/arm64") {
			pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, "windows")
		} else {
			pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, "windows/amd64")
		}
	} else if has(envs, "windows/arm64") {
		pkgInfo.SupportedEnvs = append(pkgInfo.SupportedEnvs, "windows/arm64")
	}
	if reflect.DeepEqual(pkgInfo.SupportedEnvs, []string{"darwin", "linux", "windows"}) {
		pkgInfo.SupportedEnvs = nil
	}
	if reflect.DeepEqual(pkgInfo.SupportedEnvs, []string{"darwin", "linux", "windows/amd64"}) {
		pkgInfo.SupportedEnvs = []string{"darwin", "linux", "amd64"}
	}
	if reflect.DeepEqual(pkgInfo.SupportedEnvs, []string{"darwin", "linux/amd64", "windows/amd64"}) {
		pkgInfo.SupportedEnvs = []string{"darwin", "amd64"}
	}
}
