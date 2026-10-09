package download

import "testing"

func TestForgejoReleaseURL(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		host  string
		owner string
		repo  string
		tag   string
		asset string
		exp   string
	}{
		{
			title: "normal",
			host:  "codeberg.org",
			owner: "mergiraf",
			repo:  "mergiraf",
			tag:   "v0.20.0",
			asset: "mergiraf_x86_64-unknown-linux-gnu.tar.gz",
			exp:   "https://codeberg.org/mergiraf/mergiraf/releases/download/v0.20.0/mergiraf_x86_64-unknown-linux-gnu.tar.gz",
		},
		{
			title: "a tag holding a slash",
			host:  "codeberg.org",
			owner: "an-owner",
			repo:  "a-repo",
			tag:   "kustomize/v5.8.1",
			asset: "a-repo.tar.gz",
			exp:   "https://codeberg.org/an-owner/a-repo/releases/download/kustomize%2Fv5.8.1/a-repo.tar.gz",
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if u := forgejoReleaseURL(d.host, d.owner, d.repo, d.tag, d.asset); u != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, u)
			}
		})
	}
}
