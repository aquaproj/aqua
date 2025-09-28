package asset

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// UpdateASTFile updates the packages section of a YAML AST file with new package data.
// It merges the provided packages into the existing packages array or replaces null values.
// The function handles both null and sequence types for the packages field.
func UpdateASTFile(file *ast.File, pkgs any) error { //nolint:cyclop
	node, err := yaml.ValueToNode(pkgs)
	if err != nil {
		return fmt.Errorf("convert packages to node: %w", err)
	}

	for _, doc := range file.Docs {
		body, ok := doc.Body.(*ast.MappingNode)
		if !ok {
			continue
		}
		for _, mapValue := range body.Values {
			key, ok := mapValue.Key.(*ast.StringNode)
			if !ok {
				continue
			}
			if key.Value != "packages" {
				continue
			}
			switch mapValue.Value.Type() {
			case ast.NullType:
				mapValue.Value = node
			case ast.SequenceType:
				if err := ast.Merge(mapValue.Value, node); err != nil {
					return fmt.Errorf("merge packages: %w", err)
				}
			default:
				return errors.New("packages must be null or array")
			}
			break
		}
	}
	return nil
}
