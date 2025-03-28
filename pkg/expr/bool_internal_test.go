package expr

import (
	"testing"
)

func Test_evaluateBool(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title      string
		expression string
		env        any
		input      any
		exp        bool
		isErr      bool
	}{
		{
			title:      "true",
			expression: `GOOS == "darwin"`,
			env: map[string]any{
				"GOOS":   "",
				"GOARCH": "",
			},
			input: map[string]any{
				"GOOS":   "darwin",
				"GOARCH": "arm64",
			},
			exp: true,
		},
		{
			title:      "false",
			expression: `GOOS == "linux"`,
			env: map[string]any{
				"GOOS":   "",
				"GOARCH": "",
			},
			input: map[string]any{
				"GOOS":   "darwin",
				"GOARCH": "arm64",
			},
			exp: false,
		},
		{
			title:      "error",
			expression: `GOOS == darwin`,
			env: map[string]any{
				"GOOS":   "",
				"GOARCH": "",
			},
			input: map[string]any{
				"GOOS":   "darwin",
				"GOARCH": "arm64",
			},
			isErr: true,
		},
	}

	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			b, err := evaluateBool(d.expression, d.env, d.input)
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
