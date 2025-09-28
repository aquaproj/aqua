// Package ast provides utility functions for working with YAML Abstract Syntax Trees (AST).
// It offers helper functions to navigate and manipulate YAML AST nodes, particularly
// for finding and extracting mapping values from YAML documents.
package ast

import (
	"errors"

	"github.com/goccy/go-yaml/ast"
)

// FindMappingValueFromNode searches for a mapping value with the specified key in a YAML AST node.
// It first normalizes the node to extract mapping values, then searches for the key.
// Returns the matching MappingValueNode or nil if not found.
func FindMappingValueFromNode(body ast.Node, key string) (*ast.MappingValueNode, error) {
	values, err := NormalizeMappingValueNodes(body)
	if err != nil {
		return nil, err
	}
	return findMappingValue(values, key), nil
}

// findMappingValue searches through a slice of mapping value nodes for a specific key.
// It iterates through the values and matches string keys against the target key.
// Returns the matching MappingValueNode or nil if not found.
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

// NormalizeMappingValueNodes extracts mapping value nodes from different AST node types.
// It handles both MappingNode (which contains multiple values) and single MappingValueNode.
// Returns a slice of MappingValueNodes for consistent processing, or an error for unsupported types.
func NormalizeMappingValueNodes(node ast.Node) ([]*ast.MappingValueNode, error) {
	switch t := node.(type) {
	case *ast.MappingNode:
		return t.Values, nil
	case *ast.MappingValueNode:
		return []*ast.MappingValueNode{t}, nil
	}
	return nil, errors.New("node must be a mapping node or mapping value node")
}
