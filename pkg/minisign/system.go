package minisign

import (
	"fmt"
	"os/exec"
)

// LookSystemExe searches for a minisign command installed on the system and
// returns its path. It is used as a fallback for environments where aqua
// doesn't manage a minisign binary of its own.
func LookSystemExe() (string, error) {
	p, err := exec.LookPath(pkgName)
	if err != nil {
		return "", fmt.Errorf("look for minisign in PATH: %w", err)
	}
	return p, nil
}
