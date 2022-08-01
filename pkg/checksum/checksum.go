package checksum

import "github.com/codingsince1985/checksum"

var SHA256sum func(filename string) (string, error) = checksum.SHA256sum //nolint:gochecknoglobals
