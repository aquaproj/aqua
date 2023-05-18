package unarchive

import (
	"context"
	"fmt"

	"github.com/mholt/archiver/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type unarchiverWithUnarchiver struct {
	unarchiver archiver.Unarchiver
	dest       string
	fs         afero.Fs
}

func (unarchiver *unarchiverWithUnarchiver) Unarchive(ctx context.Context, logE *logrus.Entry, src *File) error {
	tempFilePath, err := src.Body.GetPath()
	if err != nil {
		return fmt.Errorf("get a temporal file: %w", err)
	}
	return unarchiver.unarchiver.Unarchive(tempFilePath, unarchiver.dest) //nolint:wrapcheck
}
