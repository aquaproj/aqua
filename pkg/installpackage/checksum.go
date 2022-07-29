package installpackage

import (
	"github.com/spf13/afero"
)

func (inst *Installer) ReadChecksumFile(fs afero.Fs, p string) error {
	return inst.checksums.ReadFile(fs, p) //nolint:wrapcheck
}

func (inst *Installer) UpdateChecksumFile(fs afero.Fs, p string) error {
	return inst.checksums.UpdateFile(fs, p) //nolint:wrapcheck
}

// func (inst *Installer) verifyChecksum(pkg *config.Package, assetName string, body io.Reader) (io.ReadCloser, error) { //nolint:cyclop
// 	pkgInfo := pkg.PackageInfo
// 	tempDir, err := afero.TempDir(inst.fs, "", "")
// 	if err != nil {
// 		return nil, fmt.Errorf("create a temporal directory: %w", err)
// 	}
// 	defer inst.fs.RemoveAll(tempDir) //nolint:errcheck
// 	tempFilePath := filepath.Join(tempDir, assetName)
// 	if assetName == "" && (pkgInfo.Type == "github_archive" || pkgInfo.Type == "go") {
// 		tempFilePath = filepath.Join(tempDir, "archive.tar.gz")
// 	}
// 	file, err := inst.fs.Create(tempFilePath)
// 	if err != nil {
// 		return nil, fmt.Errorf("create a temporal file: %w", logerr.WithFields(err, logrus.Fields{
// 			"temp_file": tempFilePath,
// 		}))
// 	}
// 	defer file.Close()
// 	if _, err := io.Copy(file, body); err != nil {
// 		return nil, err //nolint:wrapcheck
// 	}
// 	sha256, err := checksum.SHA256sum(tempFilePath)
// 	if err != nil {
// 		return nil, fmt.Errorf("calculate a checksum of downloaded file: %w", logerr.WithFields(err, logrus.Fields{
// 			"temp_file": tempFilePath,
// 		}))
// 	}
//
// 	checksumID, err := pkg.GetChecksumID(inst.runtime)
// 	if err != nil {
// 		return nil, err //nolint:wrapcheck
// 	}
// 	chksum := inst.checksums.Get(checksumID)
//
// 	if chksum != "" && sha256 != chksum {
// 		return nil, logerr.WithFields(errInvalidChecksum, logrus.Fields{ //nolint:wrapcheck
// 			"actual_checksum":   sha256,
// 			"expected_checksum": chksum,
// 		})
// 	}
// 	if chksum == "" {
// 		inst.checksums.Set(checksumID, sha256)
// 	}
// 	readFile, err := inst.fs.Open(tempFilePath)
// 	if err != nil {
// 		return nil, err //nolint:wrapcheck
// 	}
// 	return readFile, nil
// }
