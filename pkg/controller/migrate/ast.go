package migrate

import (
	"fmt"
	"os"
	"strings"

	wast "github.com/aquaproj/aqua/v2/pkg/ast"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

// checksumKey is the section that says what aqua-checksums.json was for.
const checksumKey = "checksum"

// keptChecksumKeys are the settings that mean something to a configuration with a lock
// file. supported_envs says which environments a lock file records, which is still a
// question; the rest were about a file that is no longer read.
var keptChecksumKeys = map[string]struct{}{ //nolint:gochecknoglobals
	"supported_envs": {},
}

// checksumSettings is what the file says about aqua-checksums.json that a lock file says
// instead, as "checksum.<key>: <value>" for each.
func checksumSettings(cfgFilePath string) ([]string, error) {
	mapping, _, err := checksumMapping(cfgFilePath)
	if err != nil || mapping == nil {
		return nil, err
	}
	var settings []string
	for _, entry := range mapping.Values {
		key, ok := entry.Key.(*ast.StringNode)
		if !ok {
			continue
		}
		if _, kept := keptChecksumKeys[key.Value]; kept {
			continue
		}
		settings = append(settings, checksumKey+"."+key.Value+": "+strings.TrimSpace(entry.Value.String()))
	}
	return settings, nil
}

// removeChecksumSettings writes the file without them, and without the section when
// nothing is left of it.
//
// The file is edited as a syntax tree, because aqua.yaml is written by people: it has
// comments, an order somebody chose, and a shape the encoder wouldn't reproduce.
func removeChecksumSettings(cfgFilePath string) error {
	mapping, file, err := checksumMapping(cfgFilePath)
	if err != nil {
		return err
	}
	if mapping == nil {
		return nil
	}
	body, err := wast.NormalizeMappingValueNodes(file.Docs[0].Body)
	if err != nil {
		return fmt.Errorf("normalize a mapping value node: %w", err)
	}

	kept := make([]*ast.MappingValueNode, 0, len(mapping.Values))
	for _, entry := range mapping.Values {
		key, ok := entry.Key.(*ast.StringNode)
		if !ok {
			kept = append(kept, entry)
			continue
		}
		if _, keep := keptChecksumKeys[key.Value]; keep {
			kept = append(kept, entry)
		}
	}

	section := findValue(body, checksumKey)
	switch {
	case len(kept) == 0:
		// The section said nothing else, so it goes with them.
		if err := removeValue(file, section); err != nil {
			return err
		}
	default:
		mapping.Values = kept
		section.Value = mapping
	}
	return write(cfgFilePath, file)
}

// checksumMapping is the checksum section of the file, as a mapping, and the file it is in.
// A file with no section, or one whose section is a single setting, is not an error.
func checksumMapping(cfgFilePath string) (*ast.MappingNode, *ast.File, error) {
	b, err := os.ReadFile(cfgFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("read a configuration file: %w", err)
	}
	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("parse a configuration file as YAML: %w", err)
	}
	if len(file.Docs) == 0 {
		return nil, file, nil
	}
	values, err := wast.NormalizeMappingValueNodes(file.Docs[0].Body)
	if err != nil {
		return nil, file, fmt.Errorf("normalize a mapping value node: %w", err)
	}
	section := findValue(values, checksumKey)
	if section == nil {
		return nil, file, nil
	}
	switch t := section.Value.(type) {
	case *ast.MappingNode:
		return t, file, nil
	case *ast.MappingValueNode:
		// One setting parses as a mapping value rather than a mapping, and says the same
		// thing.
		return &ast.MappingNode{Values: []*ast.MappingValueNode{t}}, file, nil
	}
	return nil, file, nil
}

// findValue is the entry under the key, or nothing.
func findValue(values []*ast.MappingValueNode, key string) *ast.MappingValueNode {
	for _, value := range values {
		if name, ok := value.Key.(*ast.StringNode); ok && name.Value == key {
			return value
		}
	}
	return nil
}

// removeValue takes one entry out of the document's top level.
func removeValue(file *ast.File, value *ast.MappingValueNode) error {
	mapping, ok := file.Docs[0].Body.(*ast.MappingNode)
	if !ok {
		// A document of one entry is that entry, and a configuration of nothing but a
		// checksum section is not something to write back empty.
		return errOnlySection
	}
	kept := make([]*ast.MappingValueNode, 0, len(mapping.Values))
	for _, v := range mapping.Values {
		if v != value {
			kept = append(kept, v)
		}
	}
	mapping.Values = kept
	return nil
}

// write stores the file, keeping the mode it had.
func write(cfgFilePath string, file *ast.File) error {
	stat, err := os.Stat(cfgFilePath)
	if err != nil {
		return fmt.Errorf("get a configuration file stat: %w", err)
	}
	content := strings.TrimSuffix(file.String(), "\n") + "\n"
	if err := os.WriteFile(cfgFilePath, []byte(content), stat.Mode()); err != nil {
		return fmt.Errorf("write a configuration file: %w", err)
	}
	return nil
}

// cutVersion splits a package name that carries its version.
func cutVersion(name string) (string, string, bool) {
	pkgName, version, found := strings.Cut(name, "@")
	return pkgName, version, found
}
