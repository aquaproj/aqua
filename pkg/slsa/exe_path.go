package slsa

import (
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/runtime"
)

type ParamExePath struct {
	RootDir string
	Runtime *runtime.Runtime
}

func ExePath(param *ParamExePath) string {
	assetName := fmt.Sprintf("slsa-verifier-%s-%s", param.Runtime.GOOS, param.Runtime.GOARCH)
	if param.Runtime.GOOS == "windows" {
		assetName += ".exe"
	}
	return filepath.Join(param.RootDir, "pkgs", "github_release", "github.com", "slsa-framework", "slsa-verifier", Version, assetName, assetName)
}
