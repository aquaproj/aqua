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

type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

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
		return "", fmt.Errorf("copy an io.Reader to hash object: %w", err)
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
	for _, suffix := range []string{"sig", "asc", "pem", "bundle"} {
		if strings.HasSuffix(s, "."+suffix) {
			return nil
		}
	}

	arr := []struct {
		words     []string
		algorithm string
	}{
		{
			words:     []string{"sha512", "shasums512"},
			algorithm: "sha512",
		},
		{
			words:     []string{"md5"},
			algorithm: "md5",
		},
		{
			words:     []string{"sha1", "shasum1"},
			algorithm: "sha1",
		},
		{
			words:     []string{"sha256", "shasums", "checksum"},
			algorithm: "sha256",
		},
	}
	for _, a := range arr {
		for _, w := range a.words {
			if strings.Contains(s, w) {
				return &registry.Checksum{
					Type:      "github_release",
					Algorithm: a.algorithm,
					Asset:     convertChecksumFileName(filename, version),
				}
			}
		}
	}

	return nil
}
