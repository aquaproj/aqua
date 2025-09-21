package checksum

import "errors"

var (
	// errInvalidChecksum indicates that a checksum value is malformed or incorrect.
	errInvalidChecksum = errors.New("checksum is invalid")

	// errUnknownChecksumFileFormat indicates that the checksum file format is not recognized.
	errUnknownChecksumFileFormat = errors.New("checksum file format is unknown")

	// ErrNoChecksumExtracted indicates that no checksum could be extracted from the source.
	ErrNoChecksumExtracted = errors.New("no checksum is extracted")

	// ErrNoChecksumIsFound indicates that no checksum was found for the requested resource.
	ErrNoChecksumIsFound = errors.New("no checksum is found")
)
