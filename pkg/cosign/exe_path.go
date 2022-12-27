package cosign

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
	assetName := fmt.Sprintf("cosign-%s-%s", param.Runtime.GOOS, param.Runtime.GOARCH)
	p := filepath.Join(param.RootDir, "pkgs", "github_release", "github.com", "sigstore", "cosign", Version, assetName, assetName)
	if param.Runtime.GOOS == "windows" {
		p += ".exe"
	}
	return p
}
