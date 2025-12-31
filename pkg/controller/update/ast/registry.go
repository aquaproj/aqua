package ast

import (
	"errors"
	"fmt"
	"log/slog"

	wast "github.com/aquaproj/aqua/v2/pkg/ast"
	"github.com/goccy/go-yaml/ast"
)

const typeStandard = "standard"

func UpdateRegistries(logger *slog.Logger, file *ast.File, newVersions map[string]string) (bool, error) {
	body := file.Docs[0].Body // DocumentNode

	mv, err := wast.FindMappingValueFromNode(body, "registries")
	if err != nil {
		return false, fmt.Errorf(`find a mapping value node "registries": %w`, err)
	}

	seq, ok := mv.Value.(*ast.SequenceNode)
	if !ok {
		return false, errors.New("the value must be a sequence node")
	}
	updated := false
	for _, value := range seq.Values {
		up, err := parseRegistryNode(logger, value, newVersions)
		if err != nil {
			return false, err
		}
		if up {
			updated = true
		}
	}
	return updated, nil
}

func updateRegistryVersion(logger *slog.Logger, refNode *ast.StringNode, rgstName, newVersion string) bool {
	if refNode.Value == newVersion {
		return false
	}
	if commitHashPattern.MatchString(refNode.Value) {
		logger.Debug("skip updating a commit hash",
			"registry_name", rgstName,
			"current_version", refNode.Value)
		return false
	}
	logger.Info("updating a registry",
		"old_version", refNode.Value,
		"new_version", newVersion,
		"registry_name", rgstName)
	refNode.Value = newVersion
	return true
}

func parseRegistryNode(logger *slog.Logger, node ast.Node, newVersions map[string]string) (bool, error) { //nolint:gocognit,cyclop,funlen
	mvs, err := wast.NormalizeMappingValueNodes(node)
	if err != nil {
		return false, fmt.Errorf("normalize a mapping value node: %w", err)
	}
	var refNode *ast.StringNode
	var newVersion string
	var rgstName string
	for _, mvn := range mvs {
		key, ok := mvn.Key.(*ast.StringNode)
		if !ok {
			continue
		}
		switch key.Value {
		case "ref":
			sn, ok := mvn.Value.(*ast.StringNode)
			if !ok {
				return false, errors.New("ref must be a string")
			}
			if newVersion == "" {
				refNode = sn
				continue
			}
			return updateRegistryVersion(logger, sn, rgstName, newVersion), nil
		case "type":
			sn, ok := mvn.Value.(*ast.StringNode)
			if !ok {
				return false, errors.New("type must be a string")
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
				return false, errors.New("name must be a string")
			}
			version, ok := newVersions[sn.Value]
			if !ok {
				return false, nil
			}
			if refNode == nil {
				rgstName = sn.Value
				newVersion = version
				continue
			}
			return updateRegistryVersion(logger, refNode, sn.Value, version), nil
		default:
			continue // Ignore unknown fields
		}
	}
	if refNode == nil || rgstName == "" {
		return false, nil
	}
	version, ok := newVersions[rgstName]
	if !ok {
		return false, nil
	}
	return updateRegistryVersion(logger, refNode, rgstName, version), nil
}
