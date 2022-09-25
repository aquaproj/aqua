package checksum

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/codingsince1985/checksum"
	"github.com/spf13/afero"
)

func Calculate(fs afero.Fs, filename, algorithm string) (string, error) {
	switch algorithm {
	case "md5":
		return checksum.MD5sum(filename) //nolint:wrapcheck
	case "sha256":
		return checksum.SHA256sum(filename) //nolint:wrapcheck
	case "sha512":
		return calculateSHA512(fs, filename)
	case "":
		return "", errors.New("algorithm is required")
	default:
		return "", errors.New("unsupported algorithm")
	}
}

func calculateSHA512(fs afero.Fs, filename string) (string, error) {
	byt, err := afero.ReadFile(fs, filename)
	if err != nil {
		return "", fmt.Errorf("read a file: %w", err)
	}
	sum := sha512.Sum512(byt)
	return hex.EncodeToString(sum[:]), nil
}

func convertChecksumFileName(filename, version string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(filename, version, "{{.Version}}"),
		strings.TrimPrefix(version, "v"), "{{trimV .Version}}")
}

func GetChecksumConfigFromFilename(filename, version string) *registry.Checksum {
	s := strings.ToLower(filename)
	for _, suffix := range []string{"sig", "asc"} {
		if strings.HasSuffix(s, "."+suffix) {
			return nil
		}
	}
	if strings.Contains(s, "sha512") {
		return &registry.Checksum{
			Type:       "github_release",
			FileFormat: "regexp",
			Algorithm:  "sha512",
			Asset:      convertChecksumFileName(filename, version),
			Pattern: &registry.ChecksumPattern{
				Checksum: `^(\b[A-Fa-f0-9]{128}\b)`,
				File:     `^\b[A-Fa-f0-9]{128}\b\s+(\S+)$`,
			},
		}
	}
	if strings.Contains(s, "sha256") || strings.Contains(s, "checksum") {
		return &registry.Checksum{
			Type:       "github_release",
			FileFormat: "regexp",
			Algorithm:  "sha256",
			Asset:      convertChecksumFileName(filename, version),
			Pattern: &registry.ChecksumPattern{
				Checksum: `^(\b[A-Fa-f0-9]{64}\b)`,
				File:     `^\b[A-Fa-f0-9]{64}\b\s+(\S+)$`,
			},
		}
	}
	return nil
}
