package unarchive

import (
	"fmt"
	"io"
	"os"

	"github.com/mholt/archiver/v3"
)

type unarchiverWithUnarchiver struct {
	unarchiver archiver.Unarchiver
	dest       string
}

func (unarchiver *unarchiverWithUnarchiver) Unarchive(body io.Reader) error {
	dest := unarchiver.dest
	f, err := os.CreateTemp("", "")
	if err != nil {
		return fmt.Errorf("create a temporal file: %w", err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	if _, err := io.Copy(f, body); err != nil {
		return fmt.Errorf("copy the file to the temporal file: %w", err)
	}
	return unarchiver.unarchiver.Unarchive(f.Name(), dest) //nolint:wrapcheck
}
