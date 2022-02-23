package controller

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

func (ctrl *Controller) generateInsert(cfgFilePath string, pkgs interface{}) error { //nolint:cyclop
	file, err := parser.ParseFile(cfgFilePath, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse configuration file as YAML: %w", err)
	}
	node, err := yaml.ValueToNode(pkgs)
	if err != nil {
		return fmt.Errorf("convert packages to node: %w", err)
	}

	changed := false
	for _, docNode := range file.Docs {
		body := docNode.Body
		mapping, ok := body.(*ast.MappingNode)
		if !ok {
			continue
		}
		for _, mv := range mapping.Values {
			if mv.Key.String() != "packages" {
				continue
			}
			seq, ok := mv.Value.(*ast.SequenceNode)
			if !ok {
				continue
			}
			if err := ast.Merge(seq, node); err != nil {
				return fmt.Errorf("merge node: %w", err)
			}
			changed = true
		}
		break
	}
	if !changed {
		return nil
	}
	stat, err := os.Stat(cfgFilePath)
	if err != nil {
		return fmt.Errorf("get configuration file stat: %w", err)
	}
	if err := os.WriteFile(cfgFilePath, []byte(file.String()), stat.Mode()); err != nil {
		return fmt.Errorf("write the configuration file: %w", err)
	}
	return nil
}
