package checksum

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/spf13/afero"
)

type Checksums struct {
	m       map[string]string
	rwmutex *sync.RWMutex
	changed bool
}

func New() *Checksums {
	return &Checksums{
		m:       map[string]string{},
		rwmutex: &sync.RWMutex{},
	}
}

func (chksums *Checksums) Get(key string) string {
	chksums.rwmutex.RLock()
	id := chksums.m[key]
	chksums.rwmutex.RUnlock()
	return id
}

func (chksums *Checksums) Set(key, value string) {
	chksums.rwmutex.Lock()
	chksums.m[key] = value
	chksums.changed = true
	chksums.rwmutex.Unlock()
}

type checksumsJSON struct {
	Checksums map[string]string `json:"checksums"`
}

func (chksums *Checksums) ReadFile(fs afero.Fs, p string) error {
	if f, err := afero.Exists(fs, p); err != nil {
		return fmt.Errorf("check if checksum file exists: %w", err)
	} else if !f {
		return nil
	}
	f, err := fs.Open(p)
	if err != nil {
		return fmt.Errorf("open a checksum file: %w", err)
	}
	defer f.Close()
	chkJSON := &checksumsJSON{}
	if err := json.NewDecoder(f).Decode(chkJSON); err != nil {
		return fmt.Errorf("parse a checksum file as JSON: %w", err)
	}
	chksums.m = chkJSON.Checksums
	return nil
}

func (chksums *Checksums) UpdateFile(fs afero.Fs, p string) error {
	if !chksums.changed {
		return nil
	}
	f, err := fs.Create(p)
	if err != nil {
		return fmt.Errorf("create a checksum file: %w", err)
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	chkJSON := &checksumsJSON{
		Checksums: chksums.m,
	}
	if err := encoder.Encode(chkJSON); err != nil {
		return fmt.Errorf("write a checksum file as JSON: %w", err)
	}
	return nil
}

func GetChecksumFilePathFromConfigFilePath(cfgFilePath string) string {
	return filepath.Join(filepath.Dir(cfgFilePath), ".aqua-checksums.json")
}
