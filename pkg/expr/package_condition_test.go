package expr_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/expr"
	"github.com/aquaproj/aqua/pkg/runtime"
)

func TestEvaluateSupportedIf(t *testing.T) {
	t.Parallel()
	data := []struct {
		title       string
		supportedIf string
		rt          *runtime.Runtime
		exp         bool
		isErr       bool
	}{
		{
			title:       "true",
			supportedIf: `GOOS == "linux"`,
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			exp: true,
		},
	}

	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			b, err := expr.EvaluateSupportedIf(&d.supportedIf, d.rt)
			if d.isErr {
				if err == nil {
					t.Fatal("err should be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if b != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, b)
			}
		})
	}
}
