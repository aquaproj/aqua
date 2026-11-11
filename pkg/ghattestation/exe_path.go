package ghattestation

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

// ParamExePath is where to look and for which machine.
type ParamExePath struct {
	RootDir string
	Runtime *runtime.Runtime
}

// ExePath is where the GitHub CLI aqua installs for attestations lives.
//
// It is derived from the package rather than written out, so that changing how the
// package is described can't leave this pointing somewhere else. cosign and
// slsa-verifier answer the same question for their own tools, and anything that
// installs one of them and then runs it needs to agree about where it went.
func ExePath(param *ParamExePath) (string, error) {
	pkg := Package()
	pkg.PackageInfo.OverrideByRuntime(param.Runtime)
	p, err := pkg.ExePath(param.RootDir, pkg.PackageInfo.GetFiles()[0], param.Runtime)
	if err != nil {
		return "", fmt.Errorf("get an executable file path of GitHub CLI: %w", err)
	}
	return p, nil
}
