package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
)

// writeLockedRegistry keeps a registry file whose checksum was checked at its
// LockedFilePath, with the JSON made from a YAML registry beside it.
//
// The content of a file there is fixed by the checksum in its path, so a file already in
// place is left alone: it is what would be written anyway, and on Windows a file another
// aqua is reading can't be replaced.
func writeLockedRegistry(registryFilePath string, content []byte, registryContent *registry.Config) error {
	if err := osfile.MkdirAll(filepath.Dir(registryFilePath)); err != nil {
		return fmt.Errorf("create the parent directory of the registry file: %w", err)
	}
	if !isJSON(registryFilePath) {
		b, err := json.Marshal(registryContent)
		if err != nil {
			return fmt.Errorf("encode a registry as JSON: %w", err)
		}
		if err := writeFileIfAbsent(registryFilePath+jsonSuffix, b); err != nil {
			return err
		}
	}
	return writeFileIfAbsent(registryFilePath, content)
}

func writeFileIfAbsent(path string, b []byte) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check if a registry file exists: %w", err)
	}
	return writeFileAtomically(path, b)
}

// writeFileAtomically writes b to a temporary file beside path and renames it to path, so
// that another aqua reading path never sees a file half written.
//
// A registry file is written only where there was none, or where its content is fixed by
// its path, so when the rename fails because another aqua has put the file in place in the
// meantime (on Windows a file being read can't be replaced), that file is kept.
func writeFileAtomically(path string, b []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create a temporary file: %w", err)
	}
	tmp := f.Name()
	if _, err := f.Write(b); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("write a temporary file: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("close a temporary file: %w", err)
	}
	if err := os.Chmod(tmp, registryFilePermission); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("change the mode of a temporary file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		if _, statErr := os.Stat(path); statErr == nil {
			return nil
		}
		return fmt.Errorf("rename a temporary file into place: %w", err)
	}
	return nil
}
