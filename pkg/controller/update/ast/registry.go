package ast

import (
	"errors"

	"github.com/goccy/go-yaml/ast"
	"github.com/sirupsen/logrus"
)

const typeStandard = "standard"

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

func updateRegistryVersion(logE *logrus.Entry, refNode *ast.StringNode, rgstName, newVersion string) {
	if refNode.Value == newVersion {
		return
	}
	logE.WithFields(logrus.Fields{
		"old_version":   refNode.Value,
		"new_version":   newVersion,
		"registry_name": rgstName,
	}).Info("updating a registry")
	refNode.Value = newVersion
}

func parseRegistryNode(logE *logrus.Entry, node ast.Node, newVersions map[string]string) error { //nolint:gocognit,cyclop,funlen
	mvs, err := normalizeMappingValueNodes(node)
	if err != nil {
		return err
	}
	var refNode *ast.StringNode
	var newVersion string
	var rgstName string
	for _, mvn := range mvs {
		switch mvn.Key.String() {
		case "ref":
			sn, ok := mvn.Value.(*ast.StringNode)
			if !ok {
				return errors.New("ref must be a string")
			}
			if newVersion == "" {
				refNode = sn
				continue
			}
			updateRegistryVersion(logE, sn, rgstName, newVersion)
			return nil
		case "type":
			sn, ok := mvn.Value.(*ast.StringNode)
			if !ok {
				return errors.New("type must be a string")
			}
			if sn.Value != typeStandard && sn.Value != "github_content" {
				break
			}
			if sn.Value != typeStandard {
				continue
			}
			if rgstName == "" {
				rgstName = typeStandard
			}
			continue
		case "name":
			sn, ok := mvn.Value.(*ast.StringNode)
			if !ok {
				return errors.New("name must be a string")
			}
			version, ok := newVersions[sn.Value]
			if !ok {
				return nil
			}
			if refNode == nil {
				rgstName = sn.Value
				newVersion = version
				continue
			}
			updateRegistryVersion(logE, refNode, sn.Value, version)
			return nil
		default:
			continue // Ignore unknown fields
		}
	}
	if refNode == nil || rgstName == "" {
		return nil
	}
	version, ok := newVersions[rgstName]
	if !ok {
		return nil
	}
	updateRegistryVersion(logE, refNode, rgstName, version)
	return nil
}

func UpdateRegistries(logE *logrus.Entry, file *ast.File, newVersions map[string]string) error {
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
		if err := parseRegistryNode(logE, value, newVersions); err != nil {
			return err
		}
	}
	return nil
}
