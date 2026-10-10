package forge_test

import (
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/versiongetter/forge"
)

// What an instance says about a refusal is what an error carries, on one line.
func TestSaid(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		body  string
		exp   string
	}{
		{
			title: "one line, whatever the instance wrote it as",
			body:  "  {\"message\":\"The target couldn't be found.\"}\n",
			exp:   `{"message":"The target couldn't be found."}`,
		},
		{
			title: "nothing said",
			body:  "\n\n",
			exp:   "",
		},
		{
			// 300 letters of two bytes each, written as an escape because the
			// source is English: by bytes this would be cut in the middle of a
			// letter, and a message of 150 such letters would be sliced out of
			// range.
			title: "a message that doesn't write a letter in one byte",
			body:  strings.Repeat("\u00e9", 300),
			exp:   strings.Repeat("\u00e9", 200) + "...",
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if s := forge.Said([]byte(d.body)); s != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, s)
			}
		})
	}
}
