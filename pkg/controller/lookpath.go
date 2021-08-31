package controller

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func lookPath(exeName, ignoredPath string) (string, error) {
	a, err := filepath.Abs(ignoredPath)
	if err != nil {
		return "", fmt.Errorf("get the absolute path (%s): %w", ignoredPath, err)
	}
	ignoredPath = a
	oldPath := os.Getenv("PATH")
	newPath, err := getNewPath(oldPath, ignoredPath)
	if err != nil {
		return "", err
	}
	os.Setenv("PATH", newPath)
	defer os.Setenv("PATH", oldPath)
	s, err := exec.LookPath(exeName)
	if err != nil {
		return "", fmt.Errorf("look the command (%s): %w", exeName, err)
	}
	return s, nil
}

func getNewPath(path, ignoredPath string) (string, error) {
	paths := strings.Split(path, ":")
	filteredPaths := []string{}
	for _, p := range paths {
		a, err := filepath.Abs(p)
		if err != nil {
			return "", fmt.Errorf("get the absolute path (%s): %w", p, err)
		}
		p = a
		if p != ignoredPath {
			filteredPaths = append(filteredPaths, p)
		}
	}
	return strings.Join(filteredPaths, ":"), nil
}
