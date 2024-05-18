package output

import (
	"errors"
	"fmt"
	"strings"

	wast "github.com/aquaproj/aqua/v2/pkg/ast"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (o *Outputter) generateInsert(cfgFilePath string, pkgs []*aqua.Package) error {
	b, err := afero.ReadFile(o.fs, cfgFilePath)
	if err != nil {
		return fmt.Errorf("read a configuration file: %w", err)
	}
	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse configuration file as YAML: %w", err)
	}

	if len(file.Docs) != 1 {
		return logerr.WithFields(errDocumentMustBeOne, logrus.Fields{ //nolint:wrapcheck
			"num_of_docs": len(file.Docs),
		})
	}
	s, err := o.getAppendedTxt(cfgFilePath, file, pkgs)
	if err != nil {
		return err
	}

	stat, err := o.fs.Stat(cfgFilePath)
	if err != nil {
		return fmt.Errorf("get configuration file stat: %w", err)
	}
	if err := afero.WriteFile(o.fs, cfgFilePath, []byte(s), stat.Mode()); err != nil {
		return fmt.Errorf("write the configuration file: %w", err)
	}
	return nil
}

func (o *Outputter) getAppendedTxt(cfgFilePath string, file *ast.File, pkgs []*aqua.Package) (string, error) {
	body := file.Docs[0].Body
	values, err := wast.FindMappingValueFromNode(body, "packages")
	if err != nil {
		return "", fmt.Errorf(`find a mapping value node "packages": %w`, err)
	}
	if values == nil {
		return o.appendPkgsTxt(cfgFilePath, pkgs)
	}

	if err := updateASTFile(values, pkgs); err != nil {
		return "", err
	}
	return file.String(), nil
}

func appendPkgsNode(mapValue *ast.MappingValueNode, node ast.Node) error {
	switch mapValue.Value.Type() {
	case ast.NullType:
		mapValue.Value = node
		return nil
	case ast.SequenceType:
		if err := ast.Merge(mapValue.Value, node); err != nil {
			return fmt.Errorf("merge packages: %w", err)
		}
		return nil
	default:
		return errors.New("packages must be null or array")
	}
}

func updateASTFile(values *ast.MappingValueNode, pkgs []*aqua.Package) error {
	node, err := yaml.ValueToNode(pkgs)
	if err != nil {
		return fmt.Errorf("convert packages to node: %w", err)
	}

	return appendPkgsNode(values, node)
}

func (o *Outputter) appendPkgsTxt(cfgFilePath string, pkgs []*aqua.Package) (string, error) {
	a, err := yaml.Marshal(struct {
		Packages []*aqua.Package `yaml:"packages"`
	}{
		Packages: pkgs,
	})
	if err != nil {
		return "", fmt.Errorf("marshal packages: %w", err)
	}
	b, err := afero.ReadFile(o.fs, cfgFilePath)
	if err != nil {
		return "", fmt.Errorf("read a configuration file: %w", err)
	}
	sb := string(b)
	if !strings.HasSuffix(sb, "\n") {
		sb += "\n"
	}
	return sb + string(a), nil
}
