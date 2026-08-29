package ghattestation

import "testing"

func Test_unescapeSignerWorkflow(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		s    string
		exp  string
	}{
		{
			name: "escaped dots, as registries wrote them for gh < v2.97.0",
			s:    `pnpm/pnpm/\.github/workflows/release\.yml`,
			exp:  "pnpm/pnpm/.github/workflows/release.yml",
		},
		{
			name: "already literal, left alone",
			s:    "pnpm/pnpm/.github/workflows/release.yml",
			exp:  "pnpm/pnpm/.github/workflows/release.yml",
		},
		{
			name: "empty",
			s:    "",
			exp:  "",
		},
		{
			name: "other escaped metacharacters",
			s:    `owner/repo/\.github/workflows/release\+build\(1\)\.yml`,
			exp:  "owner/repo/.github/workflows/release+build(1).yml",
		},
		{
			name: "escaped backslash becomes one backslash",
			s:    `owner/repo/a\\b`,
			exp:  `owner/repo/a\b`,
		},
		{
			name: "unescaped metacharacters are not touched",
			s:    "owner/repo/.*",
			exp:  "owner/repo/.*",
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			if got := unescapeSignerWorkflow(d.s); got != d.exp {
				t.Fatalf("got %q, wanted %q", got, d.exp)
			}
		})
	}
}
