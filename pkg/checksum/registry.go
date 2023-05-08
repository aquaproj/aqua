package checksum

import (
	"fmt"
	"path"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func RegistryID(regist *aqua.Registry) string {
	return path.Join("registries", "github_content", "github.com", regist.RepoOwner, regist.RepoName, regist.Ref, regist.Path)
}

func CheckRegistry(regist *aqua.Registry, checksums *Checksums, content []byte) error {
	checksumID := RegistryID(regist)
	chksum := checksums.Get(checksumID)
	algorithm := "sha512"
	if chksum != nil {
		algorithm = chksum.Algorithm
	}
	chk, err := CalculateReader(strings.NewReader(string(content)), algorithm)
	if err != nil {
		return fmt.Errorf("calculate a checksum: %w", err)
	}
	if chksum == nil {
		chksum = &Checksum{
			ID:        checksumID,
			Algorithm: "sha512",
			Checksum:  chk,
		}
		checksums.Set(checksumID, chksum)
		return nil
	}
	if !strings.EqualFold(chksum.Checksum, chk) {
		return logerr.WithFields(errInvalidChecksum, logrus.Fields{ //nolint:wrapcheck
			"actual_checksum":   strings.ToUpper(chk),
			"expected_checksum": strings.ToUpper(chksum.Checksum),
		})
	}
	return nil
}
