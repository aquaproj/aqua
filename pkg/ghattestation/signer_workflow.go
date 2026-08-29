package ghattestation

import "regexp"

// regexpEscape matches a backslash escaping a regular expression metacharacter.
var regexpEscape = regexp.MustCompile(`\\([.\\+*?()|\[\]{}^$])`)

// unescapeSignerWorkflow turns a regex-style signer workflow into a literal path.
//
// gh < v2.97.0 used --signer-workflow as a regular expression, so registries
// escaped the dots in the path. v2.97.0 passes the value through
// regexp.QuoteMeta and matches it literally, which escapes those backslashes
// again and makes the matcher require a backslash that no certificate SAN has.
//
// Registry entries written for the old behaviour would therefore stop
// verifying the moment aqua bumped its bundled gh. Undoing the escaping keeps
// them working. Values that were already literal have no backslashes and pass
// through unchanged.
//
// TODO: remove once registries pinning the escaped form are old enough to drop.
func unescapeSignerWorkflow(s string) string {
	return regexpEscape.ReplaceAllString(s, "$1")
}
