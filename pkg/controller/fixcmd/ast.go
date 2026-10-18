package fixcmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	wast "github.com/aquaproj/aqua/v2/pkg/ast"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

// renameInFile rewrites the names of the packages the map names, and reports whether the
// file held any of them.
//
// The file is edited as a syntax tree rather than read into a Config and written back
// out. aqua.yaml is written by people: it has comments, an order somebody chose, and a
// shape the encoder wouldn't reproduce, and none of that is this command's to decide.
// What it changes is one string per package.
func renameInFile(logger *slog.Logger, cfgFilePath string, renames map[string]string) (bool, error) {
	b, err := os.ReadFile(cfgFilePath)
	if err != nil {
		return false, fmt.Errorf("read a configuration file: %w", err)
	}
	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return false, fmt.Errorf("parse a configuration file as YAML: %w", err)
	}

	renamed, err := renamePackages(logger, file, renames)
	if err != nil {
		return false, err
	}
	if !renamed {
		return false, nil
	}

	stat, err := os.Stat(cfgFilePath)
	if err != nil {
		return false, fmt.Errorf("get a configuration file stat: %w", err)
	}
	if err := os.WriteFile(cfgFilePath, []byte(file.String()), stat.Mode()); err != nil { //nolint:gosec // the path is the configuration file aqua was pointed at
		return false, fmt.Errorf("write a configuration file: %w", err)
	}
	return true, nil
}

// renamePackages rewrites the name of every package the map names.
func renamePackages(logger *slog.Logger, file *ast.File, renames map[string]string) (bool, error) {
	if len(file.Docs) == 0 {
		return false, nil
	}
	mv, err := wast.FindMappingValueFromNode(file.Docs[0].Body, "packages")
	if err != nil {
		return false, fmt.Errorf(`find a mapping value node "packages": %w`, err)
	}
	if mv == nil {
		// A configuration that declares no packages, which is what a file holding only
		// registries or imports looks like.
		return false, nil
	}
	seq, ok := mv.Value.(*ast.SequenceNode)
	if !ok {
		return false, errors.New("packages must be a sequence")
	}

	renamed := false
	for _, value := range seq.Values {
		ok, err := renamePackage(logger, value, renames)
		if err != nil {
			return false, err
		}
		renamed = renamed || ok
	}
	return renamed, nil
}

// renamePackage rewrites one package's name, and reports whether it did.
func renamePackage(logger *slog.Logger, node ast.Node, renames map[string]string) (bool, error) {
	mvs, err := wast.NormalizeMappingValueNodes(node)
	if err != nil {
		return false, fmt.Errorf("normalize a mapping value node: %w", err)
	}
	nameNode, registryName := packageNode(mvs)
	if nameNode == nil {
		// An entry with no name is an import, or a package the file says nothing about
		// that this could act on.
		return false, nil
	}
	if registryName != "" && registryName != standardRegistry {
		// Read again here rather than trusted from the map, which is keyed by name: two
		// registries can hold a package of the same name, and only this one's aliases
		// are what was resolved.
		return false, nil
	}

	name, version, hasVersion := strings.Cut(nameNode.Value, "@")
	to, ok := renames[name]
	if !ok {
		return false, nil
	}
	if hasVersion {
		// The version stays exactly as it was. What changes is what the package is
		// called, not which release of it is installed.
		nameNode.Value = to + "@" + version
	} else {
		nameNode.Value = to
	}
	logger.Info("renaming a package", "old_name", name, "new_name", to)
	return true, nil
}

// packageNode is a package entry's name node and the registry it names, if any.
func packageNode(mvs []*ast.MappingValueNode) (*ast.StringNode, string) {
	var nameNode *ast.StringNode
	registryName := ""
	for _, mvn := range mvs {
		key, ok := mvn.Key.(*ast.StringNode)
		if !ok {
			continue
		}
		sn, ok := mvn.Value.(*ast.StringNode)
		if !ok {
			// A name or a registry that isn't a string is something this doesn't
			// understand, and an entry it leaves alone.
			continue
		}
		switch key.Value {
		case "name":
			nameNode = sn
		case "registry":
			registryName = sn.Value
		}
	}
	return nameNode, registryName
}
