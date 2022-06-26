package expr

import (
	"github.com/clivm/clivm/pkg/runtime"
)

func EvaluateSupportedIf(supportedIf *string, rt *runtime.Runtime) (bool, error) {
	if supportedIf == nil {
		return true, nil
	}
	return evaluateBool(*supportedIf, map[string]interface{}{
		"GOOS":   "",
		"GOARCH": "",
	}, map[string]interface{}{
		"GOOS":   rt.GOOS,
		"GOARCH": rt.GOARCH,
	})
}
