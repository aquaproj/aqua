package config

import (
	"testing"
)

func Test_replace(t *testing.T) {
	t.Parallel()
	data := []struct {
		title        string
		key          string
		replacements map[string]string
		exp          string
	}{
		{
			title: "replace",
			key:   "darwin",
			exp:   "x86_64",
			replacements: map[string]string{
				"darwin": "x86_64",
			},
		},
		{
			title:        "not replace",
			key:          "darwin",
			exp:          "darwin",
			replacements: map[string]string{},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			val := replace(d.key, d.replacements)
			if val != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, val)
			}
		})
	}
}
