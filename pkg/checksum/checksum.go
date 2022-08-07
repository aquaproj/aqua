package checksum

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/codingsince1985/checksum"
	"github.com/spf13/afero"
)

func Calculate(fs afero.Fs, filename, algorithm string) (string, error) {
	switch algorithm {
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
