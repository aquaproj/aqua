package checksum

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// Checksums manages a collection of checksum data with concurrent access support.
// It maintains both current and new checksum maps for efficient pruning and updates,
// tracks changes for file synchronization, and provides thread-safe operations.
type Checksums struct {
	m       map[string]*Checksum
	newM    map[string]*Checksum
	rwmutex *sync.RWMutex
	changed bool
	stdout  io.Writer
}

// New creates and initializes a new Checksums instance.
// It sets up empty maps for current and new checksums and initializes the mutex.
func New() *Checksums {
	return &Checksums{
		m:       map[string]*Checksum{},
		newM:    map[string]*Checksum{},
		rwmutex: &sync.RWMutex{},
	}
}

// Open loads checksums from a file and returns a cleanup function.
// If checksum validation is disabled, it returns nil. Otherwise, it reads
// the checksum file and provides a cleanup function to save changes on exit.
func Open(logger *slog.Logger, fs afero.Fs, cfgFilePath string, enabled bool) (*Checksums, func(), error) {
	if !enabled {
		return nil, func() {}, nil
	}
	checksumFilePath, err := GetChecksumFilePathFromConfigFilePath(fs, cfgFilePath)
	if err != nil {
		return nil, nil, err
	}
	checksums := New()
	if err := checksums.ReadFile(fs, checksumFilePath); err != nil {
		return nil, nil, fmt.Errorf("read a checksum JSON: %w", err)
	}
	return checksums, func() {
		if err := checksums.UpdateFile(fs, checksumFilePath); err != nil {
			slogerr.WithError(logger, err).Error("update a checksum file")
		}
	}, nil
}

// EnableOutput enables stdout output for checksum file updates.
// This allows the checksum manager to print file paths when updating checksum files.
func (c *Checksums) EnableOutput() {
	c.stdout = os.Stdout
}

// Get retrieves a checksum by key and marks it as accessed.
// Thread-safe operation that also adds the checksum to the new map
// to indicate it's still in use (for pruning purposes).
func (c *Checksums) Get(key string) *Checksum {
	c.rwmutex.Lock()
	chk := c.m[key]
	if chk != nil {
		c.newM[key] = chk
	}
	c.rwmutex.Unlock()
	return chk
}

// Set stores a checksum with the given key, normalizing the checksum to uppercase.
// Thread-safe operation that marks the collection as changed if it's a new entry.
// Also adds the checksum to the new map for tracking active checksums.
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

// Prune removes unused checksums by replacing the main map with the new map.
// This operation keeps only checksums that have been accessed since the last prune.
// Thread-safe operation that marks the collection as changed if items were removed.
func (c *Checksums) Prune() {
	c.rwmutex.Lock()
	if len(c.m) != len(c.newM) {
		c.changed = true
	}
	c.m = c.newM
	c.rwmutex.Unlock()
}

// checksumsJSON represents the JSON structure for serializing checksums to file.
// It wraps the checksum array in a JSON object for better extensibility.
type checksumsJSON struct {
	Checksums []*Checksum `json:"checksums"`
}

// Checksum represents a single checksum entry with its identifier, hash value, and algorithm.
// ID uniquely identifies the resource, Checksum contains the hash value,
// and Algorithm specifies the hashing method (e.g., "sha256", "sha512").
type Checksum struct {
	ID        string `json:"id"`
	Checksum  string `json:"checksum"`
	Algorithm string `json:"algorithm"`
}

// UnmarshalJSON implements json.Unmarshaler to load checksums from JSON data.
// It parses the JSON structure and populates the internal checksum map.
func (c *Checksums) UnmarshalJSON(b []byte) error {
	chkJSON := &checksumsJSON{}
	if err := json.Unmarshal(b, chkJSON); err != nil {
		return fmt.Errorf("parse a checksum file as JSON: %w", err)
	}
	m := make(map[string]*Checksum, len(chkJSON.Checksums))
	for _, chk := range chkJSON.Checksums {
		m[chk.ID] = chk
	}
	c.m = m
	return nil
}

// ReadFile loads checksums from a JSON file.
// If the file doesn't exist, it returns without error (empty state).
// Otherwise, it reads and parses the JSON content into the checksum collection.
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
	if err := json.NewDecoder(f).Decode(c); err != nil {
		return fmt.Errorf("parse a checksum file as JSON: %w", err)
	}
	return nil
}

// UpdateFile saves the current checksums to a JSON file if changes were made.
// It sorts checksums by ID for consistent output and optionally prints the file path.
// The operation is skipped if no changes have been made since the last save.
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

// GetChecksumFilePathFromConfigFilePath determines the checksum file path from a config file path.
// It checks for both "aqua-checksums.json" and ".aqua-checksums.json" in the config directory,
// preferring the non-hidden version. If neither exists, it returns the non-hidden path for creation.
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
