package controller

import (
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/goccy/go-yaml/printer"
)

func (ctrl *Controller) generateInsert(cfgFilePath string, pkgs interface{}) error {
	file, err := parser.ParseFile(cfgFilePath, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse configuration file as YAML: %w", err)
	}

	if err := ctrl.updateASTFile(file, pkgs); err != nil {
		return err
	}

	stat, err := os.Stat(cfgFilePath)
	if err != nil {
		return fmt.Errorf("get configuration file stat: %w", err)
	}
	var p printer.Printer
	if err := os.WriteFile(cfgFilePath, p.PrintNode(file.Docs[0]), stat.Mode()); err != nil {
		return fmt.Errorf("write the configuration file: %w", err)
	}
	return nil
}

func (ctrl *Controller) updateASTFile(file *ast.File, pkgs interface{}) error {
	node, err := yaml.ValueToNode(pkgs)
	if err != nil {
		return fmt.Errorf("convert packages to node: %w", err)
	}
	path, err := yaml.PathString("$.packages")
	if err != nil {
		return fmt.Errorf("build a YAML Path: %w", err)
	}
	base, err := path.FilterFile(file)
	if err != nil {
		if errors.Is(err, yaml.ErrNotFoundNode) {
			if err := path.ReplaceWithNode(file, node); err != nil {
				return fmt.Errorf("replace node: %w", err)
			}
			return nil
		}
		return fmt.Errorf("get packages with YAML Path: %w", err)
	}
	if base.Type() == ast.NullType {
		ctrl.logE().Info("replace null node")
		if err := path.ReplaceWithNode(file, node); err != nil {
			return fmt.Errorf("replace node: %w", err)
		}
		return nil
	}
	if err := ast.Merge(base, node); err != nil {
		return fmt.Errorf("add packages to AST: %w", err)
	}
	return nil
}
