package ast

import (
	"errors"

	"github.com/goccy/go-yaml/ast"
)

func FindMappingValueFromNode(body ast.Node, key string) (*ast.MappingValueNode, error) {
	values, err := NormalizeMappingValueNodes(body)
	if err != nil {
		return nil, err
	}

	return findMappingValue(values, key), nil
}

func findMappingValue(values []*ast.MappingValueNode, key string) *ast.MappingValueNode {
	for _, value := range values {
		sn, ok := value.Key.(*ast.StringNode)
		if !ok {
			continue
		}

		if sn.Value == key {
			return value
		}
	}

	return nil
}

func NormalizeMappingValueNodes(node ast.Node) ([]*ast.MappingValueNode, error) {
	switch t := node.(type) {
	case *ast.MappingNode:
		return t.Values, nil
	case *ast.MappingValueNode:
		return []*ast.MappingValueNode{t}, nil
	}

	return nil, errors.New("node must be a mapping node or mapping value node")
}
