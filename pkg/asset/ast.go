package asset

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

func UpdateASTFile(file *ast.File, pkgs interface{}) error {
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
			if mapValue.Key.String() != "packages" {
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
