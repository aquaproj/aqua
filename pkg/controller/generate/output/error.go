package output

import "errors"

var (
	errDocumentMustBeOne = errors.New("the number of document in aqua.yaml must be one")
	errBodyFormat        = errors.New("fails to parse a configuration file. Format is wrong. body must be *ast.MappingNode or *ast.MappingValueNode")
	errPkgsNotFound      = errors.New("the field 'packages' isn't found")
)
