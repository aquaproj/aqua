package genrgst

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/asset"
	"github.com/goccy/go-yaml/parser"
	"github.com/spf13/afero"
)

func (c *Controller) insert(cfgFilePath string, pkgs interface{}) error {
	b, err := afero.ReadFile(c.fs, cfgFilePath)
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

	stat, err := c.fs.Stat(cfgFilePath)
	if err != nil {
		return fmt.Errorf("get configuration file stat: %w", err)
	}
	if err := afero.WriteFile(c.fs, cfgFilePath, []byte(file.String()+"\n"), stat.Mode()); err != nil {
		return fmt.Errorf("write the configuration file: %w", err)
	}
	return nil
}
