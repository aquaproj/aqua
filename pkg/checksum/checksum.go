package checksum

import (
	"crypto/sha1" //nolint:gosec
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/codingsince1985/checksum"
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
	switch algorithm {
	case "md5":
		return checksum.MD5sumReader(file) //nolint:wrapcheck
	case "sha256":
		return checksum.SHA256sumReader(file) //nolint:wrapcheck
	case "sha512":
		return calculateSHA512Reader(file)
	case "sha1":
		return calculateSHA1Reader(file)
	case "":
		return "", errors.New("algorithm is required")
	default:
		return "", errors.New("unsupported algorithm")
	}
}

func calculateSHA1Reader(file io.Reader) (string, error) {
	byt, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("read a file: %w", err)
	}
	sum := sha1.Sum(byt) //nolint:gosec
	return hex.EncodeToString(sum[:]), nil
}

func calculateSHA512Reader(file io.Reader) (string, error) {
	byt, err := io.ReadAll(file)
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
