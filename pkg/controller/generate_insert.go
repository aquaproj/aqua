package controller

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
)

func (ctrl *Controller) generateInsert(cfgFilePath string, pkgs interface{}) error {
	file, err := parser.ParseFile(cfgFilePath, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse configuration file as YAML: %w", err)
	}
	node, err := yaml.ValueToNode(pkgs)
	if err != nil {
		return fmt.Errorf("convert packages to node: %w", err)
	}
	path, err := yaml.PathString("$.packages")
	if err != nil {
		return fmt.Errorf("build a YAML Path: %w", err)
	}
	if err := path.MergeFromNode(file, node); err != nil {
		return fmt.Errorf("add packages to AST: %w", err)
	}

	stat, err := os.Stat(cfgFilePath)
	if err != nil {
		return fmt.Errorf("get configuration file stat: %w", err)
	}
	if err := os.WriteFile(cfgFilePath, []byte(file.String()+"\n"), stat.Mode()); err != nil {
		return fmt.Errorf("write the configuration file: %w", err)
	}
	return nil
}
