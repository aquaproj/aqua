package genrgst

import (
	"fmt"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/asset"
	"github.com/goccy/go-yaml/parser"
)

func (c *Controller) insert(cfgFilePath string, pkgs any) error {
	b, err := os.ReadFile(cfgFilePath)
	if err != nil {
		return fmt.Errorf("read a configuration file: %w", err)
	}
	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse configuration file as YAML: %w", err)
	}

	if err := asset.UpdateASTFile(file, pkgs); err != nil {
		return fmt.Errorf("update an AST file: %w", err)
	}

	stat, err := os.Stat(cfgFilePath)
	if err != nil {
		return fmt.Errorf("get configuration file stat: %w", err)
	}
	if err := os.WriteFile(cfgFilePath, []byte(file.String()+"\n"), stat.Mode()); err != nil { //nolint:gosec // the path is the configuration file aqua was pointed at
		return fmt.Errorf("write the configuration file: %w", err)
	}
	return nil
}
