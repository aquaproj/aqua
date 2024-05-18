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
	body := file.Docs[0].Body

	values, err := wast.FindMappingValueFromNode(body, "packages")
	if err != nil {
		return fmt.Errorf(`find a mapping value node "packages": %w`, err)
	}
	if values == nil {
		a, err := yaml.Marshal(struct {
			Packages []*aqua.Package `yaml:"packages"`
		}{
			Packages: pkgs,
		})
		if err != nil {
			return fmt.Errorf("marshal packages: %w", err)
		}
		b, err := afero.ReadFile(o.fs, cfgFilePath)
		if err != nil {
			return fmt.Errorf("read a configuration file: %w", err)
		}
		sb := string(b)
		if !strings.HasSuffix(sb, "\n") {
			sb += "\n"
		}
		sb += string(a)
		stat, err := o.fs.Stat(cfgFilePath)
		if err != nil {
			return fmt.Errorf("get configuration file stat: %w", err)
		}
		if err := afero.WriteFile(o.fs, cfgFilePath, []byte(sb), stat.Mode()); err != nil {
			return fmt.Errorf("write the configuration file: %w", err)
		}
		return nil
	}

	if err := updateASTFile(body, pkgs); err != nil {
		return err
	}

	stat, err := o.fs.Stat(cfgFilePath)
	if err != nil {
		return fmt.Errorf("get configuration file stat: %w", err)
	}
	if err := afero.WriteFile(o.fs, cfgFilePath, []byte(file.String()), stat.Mode()); err != nil {
		return fmt.Errorf("write the configuration file: %w", err)
	}
	return nil
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

func updateASTFile(body ast.Node, pkgs []*aqua.Package) error {
	node, err := yaml.ValueToNode(pkgs)
	if err != nil {
		return fmt.Errorf("convert packages to node: %w", err)
	}

	values, err := wast.FindMappingValueFromNode(body, "packages")
	if err != nil {
		return fmt.Errorf(`find a mapping value node "packages": %w`, err)
	}

	if values == nil {
		values, err := wast.NormalizeMappingValueNodes(body)
		if err != nil {
			return err
		}
		idx := len(values)
		mn, ok := node.(*ast.MappingValueNode)
		if !ok {
			return errors.New("body must be a mapping value node")
		}
		mv, ok := body.(*ast.MappingNode)
		if !ok {
			return errors.New("body must be a mapping node")
		}
		latterValues := make([]*ast.MappingValueNode, len(mv.Values[idx:]))
		copy(latterValues, mv.Values[idx:])
		mv.Values = mv.Values[:idx]
		mv.Merge(&ast.MappingNode{
			Values: []*ast.MappingValueNode{
				mn,
			},
		})
		mv.Merge(&ast.MappingNode{
			Values: latterValues,
		})
		return nil
	}

	return appendPkgsNode(values, node)
}
