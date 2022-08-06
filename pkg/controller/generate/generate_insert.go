package generate

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/spf13/afero"
)

func (ctrl *Controller) generateInsert(cfgFilePath string, pkgs interface{}) error {
	b, err := afero.ReadFile(ctrl.fs, cfgFilePath)
	if err != nil {
		return fmt.Errorf("read a configuration file: %w", err)
	}
	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse configuration file as YAML: %w", err)
	}

	if err := ctrl.updateASTFile(file, pkgs); err != nil {
		return err
	}

	stat, err := ctrl.fs.Stat(cfgFilePath)
	if err != nil {
		return fmt.Errorf("get configuration file stat: %w", err)
	}
	if err := afero.WriteFile(ctrl.fs, cfgFilePath, []byte(file.String()+"\n"), stat.Mode()); err != nil {
		return fmt.Errorf("write the configuration file: %w", err)
	}
	return nil
}

func (ctrl *Controller) updateASTFile(file *ast.File, pkgs interface{}) error { //nolint:cyclop
	node, err := yaml.ValueToNode(pkgs)
	if err != nil {
		return fmt.Errorf("convert packages to node: %w", err)
	}

	for _, doc := range file.Docs {
		var values []*ast.MappingValueNode
		switch body := doc.Body.(type) {
		case *ast.MappingNode:
			values = body.Values
		case *ast.MappingValueNode:
			values = append(values, body)
		default:
			continue
		}
		for _, mapValue := range values {
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
