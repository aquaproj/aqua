package runtime_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

func TestNew(t *testing.T) {
	t.Parallel()
	rt := runtime.New()
	if rt == nil { //nolint:staticcheck
		t.Fatal("runtime must not be nil")
	}
	if rt.GOOS == "" { //nolint:staticcheck
		t.Fatal("rt.GOOS is empty")
	}
	if rt.GOARCH == "" { //nolint:staticcheck
		t.Fatal("rt.GOARCH is empty")
	}
}
