package checksum

import (
	"crypto/md5"  //nolint:gosec
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/spf13/afero"
)

func NewCalculator() *Calculator {
	return &Calculator{}
}

type Calculator struct{}

func (*Calculator) Calculate(fs afero.Fs, filename, algorithm string) (string, error) {
	f, err := fs.Open(filename)
	if err != nil {
		return "", fmt.Errorf("open a file to calculate the checksum: %w", err)
	}
	defer f.Close()
	return CalculateReader(f, algorithm)
}

func CalculateReader(file io.Reader, algorithm string) (string, error) {
	h, err := getHash(algorithm)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(h, file); err != nil {
		return "", fmt.Errorf("copy a io.Reader to hash object: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func getHash(algorithm string) (hash.Hash, error) {
	switch algorithm {
	case "md5":
		return md5.New(), nil //nolint:gosec
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	case "sha1":
		return sha1.New(), nil //nolint:gosec
	case "":
		return nil, errors.New("algorithm is required")
	default:
		return nil, errors.New("unsupported algorithm")
	}
}

func convertChecksumFileName(filename, version string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(filename, version, "{{.Version}}"),
		strings.TrimPrefix(version, "v"), "{{trimV .Version}}")
}

func GetChecksumConfigFromFilename(filename, version string) *registry.Checksum {
	s := strings.ToLower(filename)
	for _, suffix := range []string{"sig", "asc", "pem"} {
		if strings.HasSuffix(s, "."+suffix) {
			return nil
		}
	}
	if strings.Contains(s, "sha512") {
		return &registry.Checksum{
			Type:      "github_release",
			Algorithm: "sha512",
			Asset:     convertChecksumFileName(filename, version),
		}
	}
	if strings.Contains(s, "md5") {
		return &registry.Checksum{
			Type:      "github_release",
			Algorithm: "md5",
			Asset:     convertChecksumFileName(filename, version),
		}
	}
	if strings.Contains(s, "sha1") {
		return &registry.Checksum{
			Type:      "github_release",
			Algorithm: "sha1",
			Asset:     convertChecksumFileName(filename, version),
		}
	}
	if strings.Contains(s, "sha256") || strings.Contains(s, "checksum") {
		return &registry.Checksum{
			Type:      "github_release",
			Algorithm: "sha256",
			Asset:     convertChecksumFileName(filename, version),
		}
	}
	return nil
}
