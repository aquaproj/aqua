package ast

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/sirupsen/logrus"
)

func UpdatePackages(logE *logrus.Entry, file *ast.File, newVersions map[string]string) (bool, error) {
	body := file.Docs[0].Body // DocumentNode
	mv, err := findMappingValueFromNode(body, "packages")
	if err != nil {
		return false, err
	}

	seq, ok := mv.Value.(*ast.SequenceNode)
	if !ok {
		return false, errors.New("the value must be a sequence node")
	}
	updated := false
	for _, value := range seq.Values {
		up, err := parsePackageNode(logE, value, newVersions)
		if err != nil {
			return false, err
		}
		if up {
			updated = up
		}
	}
	return updated, nil
}

func parsePackageNode(logE *logrus.Entry, node ast.Node, newVersions map[string]string) (bool, error) { //nolint:cyclop,funlen
	mvs, err := normalizeMappingValueNodes(node)
	if err != nil {
		return false, err
	}
	var registryName string
	var pkgName string
	var pkgVersion string
	var nameNode *ast.StringNode
	for _, mvn := range mvs {
		key, ok := mvn.Key.(*ast.StringNode)
		if !ok {
			continue
		}
		switch key.Value {
		case "registry":
			sn, ok := mvn.Value.(*ast.StringNode)
			if !ok {
				return false, errors.New("registry must be a string")
			}
			registryName = sn.Value
		case "name":
			sn, ok := mvn.Value.(*ast.StringNode)
			if !ok {
				return false, errors.New("name must be a string")
			}
			nameNode = sn
			name, version, ok := strings.Cut(sn.Value, "@")
			if !ok {
				continue
			}
			pkgName = name
			pkgVersion = version
		default:
			continue // Ignore unknown fields
		}
	}
	if registryName == "" {
		registryName = "standard"
	}
	if pkgName == "" {
		return false, nil
	}
	newVersion, ok := newVersions[fmt.Sprintf("%s,%s", registryName, pkgName)]
	if !ok {
		logE.Debug("version isn't found")
		return false, nil
	}
	if pkgVersion == newVersion {
		logE.Debug("already latest")
		return false, nil
	}
	if commitHashPattern.MatchString(pkgVersion) {
		logE.WithFields(logrus.Fields{
			"current_version": pkgVersion,
			"package_name":    pkgName,
		}).Debug("skip updating a commit hash")
		return false, nil
	}
	logE.WithFields(logrus.Fields{
		"old_version":  pkgVersion,
		"new_version":  newVersion,
		"package_name": pkgName,
	}).Info("updating a package")
	nameNode.Value = fmt.Sprintf("%s@%s", pkgName, newVersion)
	return true, nil
}
