package policy

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/expr"
)

func (pc *Checker) matchRegistry(rgst *aqua.Registry, rgstPolicy *Registry) (bool, error) {
	if rgst.Type != rgstPolicy.Type {
		return false, nil
	}
	if rgst.Type == "local" {
		return rgst.Path == rgstPolicy.Path, nil
	}
	if rgst.RepoOwner != rgstPolicy.RepoOwner {
		return false, nil
	}
	if rgst.RepoName != rgstPolicy.RepoName {
		return false, nil
	}
	if rgst.Path != rgstPolicy.Path {
		return false, nil
	}

	if rgstPolicy.Ref != "" {
		matched, err := expr.EvaluateVersionConstraints(rgstPolicy.Ref, rgst.Ref, rgst.Ref)
		if err != nil {
			return false, fmt.Errorf("evaluate the version constraint of registry: %w", err)
		}
		return matched, nil
	}
	return true, nil
}
