// Package forge is what a forge instance's releases are read as.
//
// A package can be on a Forgejo, a Gitea or a GitLab instance, and each answers in its own
// words: Forgejo serves the API it inherited from Gitea at /api/v1, asking for a page with
// page and limit, while GitLab's is /api/v4, asks with page and per_page, and says
// upcoming_release where the others say draft. What they have in common is what aqua reads,
// so each client translates into the release here and one version getter reads them all.
//
// The clients are in gitea and gitlab beside this package. They import this one; this one
// imports neither, so the two can differ without the other knowing.
package forge

import "strings"

// MaxPages is how many pages are read before giving up.
//
// No instance says how many pages there are, so the end of the list is an empty page. One
// that answered the same page to every request would otherwise be read forever, and 50
// pages of releases is already far more than a version is ever found in.
const MaxPages = 50

// Release is as much of a release as aqua reads. An instance says a good deal more.
type Release struct {
	TagName string
	Name    string
	Body    string
	HTMLURL string
	// Draft says the release isn't one to install from: its assets are not served to
	// anyone but the people who can publish it. GitLab has no draft and says
	// upcoming_release, a release dated in the future, which is the same answer to the
	// same question.
	Draft      bool
	Prerelease bool
}

// Said is the start of what an instance answered with, on one line, for an error to carry.
//
// It is where a rate limit, a renamed repository and a private one are told apart: each is
// a status code with a sentence beside it, and the sentence is the part that says which.
func Said(b []byte) string {
	const maxLen = 200
	// By runes rather than bytes, so that a message in a language that doesn't write a
	// letter in one byte is neither cut in the middle of one nor sliced out of range.
	r := []rune(strings.Join(strings.Fields(string(b)), " "))
	if len(r) > maxLen {
		return string(r[:maxLen]) + "..."
	}
	return string(r)
}
