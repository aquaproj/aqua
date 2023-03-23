package checksum

import "errors"

var (
	errInvalidChecksum           = errors.New("checksum is invalid")
	errUnknownChecksumFileFormat = errors.New("checksum file format is unknown")
	ErrNoChecksumExtracted       = errors.New("no checksum is extracted")
	ErrNoChecksumIsFound         = errors.New("no checksum is found")
)
