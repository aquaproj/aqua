package render

import "github.com/aquaproj/aqua/pkg/runtime"

func Replace(key string, replacements map[string]string) string {
	a := replacements[key]
	if a == "" {
		return key
	}
	return a
}

func GetArch(rosetta2 bool, replacements map[string]string, rt *runtime.Runtime) string {
	if rosetta2 && rt.GOOS == "darwin" && rt.GOARCH == "arm64" {
		// Rosetta 2
		return Replace("amd64", replacements)
	}
	return Replace(rt.GOARCH, replacements)
}
