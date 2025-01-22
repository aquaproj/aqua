package vacuum

import (
	"time"
)

type Mock struct {
	rootDir    string
	timestamps map[string]time.Time
	err        error
}

func NewMock(rootDir string, timestamps map[string]time.Time, err error) *Mock {
	return &Mock{
		rootDir:    rootDir,
		timestamps: timestamps,
		err:        err,
	}
}

func (m *Mock) Remove(pkgPath string) error {
	return m.err
}

func (m *Mock) Update(pkgPath string, timestamp time.Time) error {
	return m.err
}

func (m *Mock) FindAll() (map[string]time.Time, error) {
	return m.timestamps, m.err
}
