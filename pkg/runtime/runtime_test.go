package runtime_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/runtime"
)

func TestNew(t *testing.T) {
	t.Parallel()
	rt := runtime.New()
	if rt == nil {
		t.Fatal("runtime must not be nil")
	}
	if rt.GOOS == "" {
		t.Fatal("rt.GOOS is empty")
	}
	if rt.GOARCH == "" {
		t.Fatal("rt.GOARCH is empty")
	}
}
