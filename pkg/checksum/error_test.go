package checksum_test

import (
	"errors"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
)

func TestErrorVariables(t *testing.T) {
	t.Parallel()

	// Test that exported error variables exist and are not nil
	if checksum.ErrNoChecksumExtracted == nil {
		t.Error("ErrNoChecksumExtracted should not be nil")
	}

	if checksum.ErrNoChecksumIsFound == nil {
		t.Error("ErrNoChecksumIsFound should not be nil")
	}

	// Test error messages
	expectedMessages := map[error]string{
		checksum.ErrNoChecksumExtracted: "no checksum is extracted",
		checksum.ErrNoChecksumIsFound:   "no checksum is found",
	}

	for err, expectedMsg := range expectedMessages {
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
	}
}

func TestErrorComparison(t *testing.T) {
	t.Parallel()

	// Test that errors can be compared using errors.Is
	testErr1 := checksum.ErrNoChecksumExtracted
	testErr2 := checksum.ErrNoChecksumIsFound

	if !errors.Is(testErr1, checksum.ErrNoChecksumExtracted) {
		t.Error("ErrNoChecksumExtracted should match itself")
	}

	if !errors.Is(testErr2, checksum.ErrNoChecksumIsFound) {
		t.Error("ErrNoChecksumIsFound should match itself")
	}

	if errors.Is(testErr1, checksum.ErrNoChecksumIsFound) {
		t.Error("ErrNoChecksumExtracted should not match ErrNoChecksumIsFound")
	}

	if errors.Is(testErr2, checksum.ErrNoChecksumExtracted) {
		t.Error("ErrNoChecksumIsFound should not match ErrNoChecksumExtracted")
	}
}

func TestErrorWrapping(t *testing.T) {
	t.Parallel()

	// Test that the errors can be wrapped and still be detected
	wrappedErr1 := errors.Join(checksum.ErrNoChecksumExtracted, errors.New("additional context"))
	wrappedErr2 := errors.Join(checksum.ErrNoChecksumIsFound, errors.New("additional context"))

	if !errors.Is(wrappedErr1, checksum.ErrNoChecksumExtracted) {
		t.Error("Wrapped ErrNoChecksumExtracted should still be detectable")
	}

	if !errors.Is(wrappedErr2, checksum.ErrNoChecksumIsFound) {
		t.Error("Wrapped ErrNoChecksumIsFound should still be detectable")
	}
}
