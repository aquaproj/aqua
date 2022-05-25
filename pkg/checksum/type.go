package checksum

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/spf13/afero"
)

type Checksums interface {
	Get(key string) string
	Set(key, value string)
	ReadFile(fs afero.Fs, p string) error
	UpdateFile(fs afero.Fs, p string) error
	Changed() bool
}

type checksums struct {
	m       map[string]string
	rwmutex *sync.RWMutex
	changed bool
}

func New() Checksums {
	return &checksums{
		m:       map[string]string{},
		rwmutex: &sync.RWMutex{},
	}
}

func (chksums *checksums) Get(key string) string {
	chksums.rwmutex.RLock()
	id := chksums.m[key]
	chksums.rwmutex.RUnlock()
	return id
}

func (chksums *checksums) Set(key, value string) {
	chksums.rwmutex.Lock()
	chksums.m[key] = value
	chksums.changed = true
	chksums.rwmutex.Unlock()
}

func (chksums *checksums) SetMap(m map[string]string) {
	chksums.m = m
}

func (chksums *checksums) ReadFile(fs afero.Fs, p string) error {
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
	if err := json.NewDecoder(f).Decode(&chksums.m); err != nil {
		return fmt.Errorf("parse a checksum file as JSON: %w", err)
	}
	return nil
}

func (chksums *checksums) UpdateFile(fs afero.Fs, p string) error {
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
	if err := encoder.Encode(chksums.m); err != nil {
		return fmt.Errorf("write a checksum file as JSON: %w", err)
	}
	return nil
}

func (chksums *checksums) Changed() bool {
	return chksums.changed
}

func GetChecksumFilePathFromConfigFilePath(cfgFilePath string) string {
	return filepath.Join(filepath.Dir(cfgFilePath), ".aqua-checksums.json")
}
