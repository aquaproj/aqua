package unarchive_test

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/unarchive"
)

// noFile stands in for a downloaded asset. The formats under test refuse before
// touching it, so it never has to be one.
type noFile struct{}

func (noFile) Path() (string, error) { return "", nil }
func (noFile) ReadLast() (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}
func (noFile) Wrap(w io.Writer) io.Writer { return w }

// dmg and pkg are opened by hdiutil and pkgutil. A caller that can't run them builds
// an Unarchiver without an executor, and that has to be refused rather than followed
// into a nil dereference.
func TestUnarchive_noExecutor(t *testing.T) {
	t.Parallel()
	for _, format := range []string{unarchive.FormatDMG, unarchive.FormatPKG} {
		t.Run(format, func(t *testing.T) {
			t.Parallel()
			err := unarchive.New(nil).Unarchive(t.Context(), slog.New(slog.DiscardHandler), &unarchive.File{
				Body:     noFile{},
				Filename: "asset." + format,
				Type:     format,
			}, t.TempDir())
			if err == nil {
				t.Fatal("an error must be returned")
			}
		})
	}
}
