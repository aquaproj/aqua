package ast

import (
	"errors"

	"github.com/goccy/go-yaml/ast"
)

func findMappingValue(values []*ast.MappingValueNode, key string) *ast.MappingValueNode {
	for _, value := range values {
		if value.Key.String() == key {
			return value
		}
	}
	return nil
}

func normalizeMappingValueNodes(node ast.Node) ([]*ast.MappingValueNode, error) {
	switch t := node.(type) {
	case *ast.MappingNode:
		return t.Values, nil
	case *ast.MappingValueNode:
		return []*ast.MappingValueNode{t}, nil
	}
	return nil, errors.New("node must be a mapping node or mapping value node")
}

func findMappingValueFromNode(body ast.Node, key string) (*ast.MappingValueNode, error) {
	values, err := normalizeMappingValueNodes(body)
	if err != nil {
		return nil, err
	}
	return findMappingValue(values, key), nil
}

func parseRegistryNode(node ast.Node, newVersions map[string]string) error { //nolint:gocognit,cyclop
	mvs, err := normalizeMappingValueNodes(node)
	if err != nil {
		return err
	}
	var refNode *ast.StringNode
	var newVersion string
	for _, mvn := range mvs {
		switch mvn.Key.String() {
		case "ref":
			sn, ok := mvn.Value.(*ast.StringNode)
			if !ok {
				return errors.New("ref must be a string")
			}
			if newVersion != "" {
				sn.Value = newVersion
				return nil
			}
			refNode = sn
		case "type":
			sn, ok := mvn.Value.(*ast.StringNode)
			if !ok {
				return errors.New("type must be a string")
			}
			if sn.Value == "standard" {
				version, ok := newVersions["standard"]
				if !ok {
					return nil
				}
				newVersion = version
			}
			if sn.Value != "standard" && sn.Value != "github_content" {
				break
			}
		case "name":
			sn, ok := mvn.Value.(*ast.StringNode)
			if !ok {
				return errors.New("name must be a string")
			}
			version, ok := newVersions[sn.Value]
			if !ok {
				return nil
			}
			if refNode != nil {
				refNode.Value = version
				return nil
			}
			newVersion = version
		default:
			continue // Ignore unknown fields
		}
	}
	return nil
}

func UpdateRegistries(file *ast.File, newVersions map[string]string) error {
	body := file.Docs[0].Body // DocumentNode
	mv, err := findMappingValueFromNode(body, "registries")
	if err != nil {
		return err
	}

	seq, ok := mv.Value.(*ast.SequenceNode)
	if !ok {
		return errors.New("the value must be a sequence node")
	}
	for _, value := range seq.Values {
		if err := parseRegistryNode(value, newVersions); err != nil {
			return err
		}
	}
	return nil
}
