package checksum

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"sync"

	"github.com/spf13/afero"
)

type Checksums struct {
	m       map[string]*Checksum
	newM    map[string]*Checksum
	rwmutex *sync.RWMutex
	changed bool
}

func New() *Checksums {
	return &Checksums{
		m:       map[string]*Checksum{},
		newM:    map[string]*Checksum{},
		rwmutex: &sync.RWMutex{},
	}
}

func (chksums *Checksums) Get(key string) *Checksum {
	chksums.rwmutex.RLock()
	chk := chksums.m[key]
	if chk != nil {
		chksums.newM[key] = chk
	}
	chksums.rwmutex.RUnlock()
	return chk
}

func (chksums *Checksums) Set(key string, chk *Checksum) {
	chksums.rwmutex.Lock()
	if _, ok := chksums.m[key]; !ok {
		chksums.m[key] = chk
		chksums.changed = true
	}
	if _, ok := chksums.newM[key]; !ok {
		chksums.newM[key] = chk
	}
	chksums.rwmutex.Unlock()
}

func (chksums *Checksums) Prune() {
	chksums.rwmutex.Lock()
	if len(chksums.m) != len(chksums.newM) {
		chksums.changed = true
	}
	chksums.m = chksums.newM
	chksums.rwmutex.Unlock()
}

type checksumsJSON struct {
	Checksums []*Checksum `json:"checksums"`
}

type Checksum struct {
	ID        string `json:"id"`
	Checksum  string `json:"checksum"`
	Algorithm string `json:"algorithm"`
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
	m := make(map[string]*Checksum, len(chkJSON.Checksums))
	for _, chk := range chkJSON.Checksums {
		m[chk.ID] = chk
	}
	chksums.m = m
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
	arr := make([]*Checksum, 0, len(chksums.m))
	for _, chk := range chksums.m {
		arr = append(arr, chk)
	}
	sort.Slice(arr, func(i, j int) bool {
		return arr[i].ID < arr[j].ID
	})
	chkJSON := &checksumsJSON{
		Checksums: arr,
	}
	if err := encoder.Encode(chkJSON); err != nil {
		return fmt.Errorf("write a checksum file as JSON: %w", err)
	}
	return nil
}

func GetChecksumFilePathFromConfigFilePath(fs afero.Fs, cfgFilePath string) (string, error) {
	p1 := filepath.Join(filepath.Dir(cfgFilePath), "aqua-checksums.json")
	f, err := afero.Exists(fs, p1)
	if err != nil {
		return "", fmt.Errorf("check if checksum file exists: %w", err)
	}
	if f {
		return p1, nil
	}

	p2 := filepath.Join(filepath.Dir(cfgFilePath), ".aqua-checksums.json")
	f, err = afero.Exists(fs, p2)
	if err != nil {
		return "", fmt.Errorf("check if checksum file exists: %w", err)
	}
	if f {
		return p2, nil
	}

	return p1, nil
}
