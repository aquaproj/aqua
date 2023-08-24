package checksum

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/afero"
)

type Checksums struct {
	m       map[string]*Checksum
	newM    map[string]*Checksum
	rwmutex *sync.RWMutex
	changed bool
	stdout  io.Writer
}

func New() *Checksums {
	return &Checksums{
		m:       map[string]*Checksum{},
		newM:    map[string]*Checksum{},
		rwmutex: &sync.RWMutex{},
	}
}

func (c *Checksums) EnableOutput() {
	c.stdout = os.Stdout
}

func (c *Checksums) Get(key string) *Checksum {
	c.rwmutex.Lock()
	chk := c.m[key]
	if chk != nil {
		c.newM[key] = chk
	}
	c.rwmutex.Unlock()
	return chk
}

func (c *Checksums) Set(key string, chk *Checksum) {
	chk.Checksum = strings.ToUpper(chk.Checksum)
	c.rwmutex.Lock()
	if _, ok := c.m[key]; !ok {
		c.m[key] = chk
		c.changed = true
	}
	if _, ok := c.newM[key]; !ok {
		c.newM[key] = chk
	}
	c.rwmutex.Unlock()
}

func (c *Checksums) Prune() {
	c.rwmutex.Lock()
	if len(c.m) != len(c.newM) {
		c.changed = true
	}
	c.m = c.newM
	c.rwmutex.Unlock()
}

type checksumsJSON struct {
	Checksums []*Checksum `json:"checksums"`
}

type Checksum struct {
	ID        string `json:"id"`
	Checksum  string `json:"checksum"`
	Algorithm string `json:"algorithm"`
}

func (c *Checksums) ReadFile(fs afero.Fs, p string) error {
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
	c.m = m
	return nil
}

func (c *Checksums) UpdateFile(fs afero.Fs, p string) error {
	if !c.changed {
		return nil
	}
	f, err := fs.Create(p)
	if err != nil {
		return fmt.Errorf("create a checksum file: %w", err)
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	arr := make([]*Checksum, 0, len(c.m))
	for _, chk := range c.m {
		arr = append(arr, chk)
	}
	sort.Slice(arr, func(i, j int) bool {
		return arr[i].ID < arr[j].ID
	})
	chkJSON := &checksumsJSON{
		Checksums: arr,
	}
	if c.stdout != nil {
		fmt.Fprintln(c.stdout, p)
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
