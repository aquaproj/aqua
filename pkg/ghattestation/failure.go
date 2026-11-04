package ghattestation

import (
	"regexp"
	"strings"
)

// failure is a kind of failures of gh attestation verify.
// GitHub CLI always exits with 1, so failures are distinguished by the output.
type failure int

const (
	// failureNone means the verification succeeded.
	failureNone failure = iota
	// failureTemporary is a temporary failure such as network errors and HTTP 5xx.
	// GitHub CLI itself retries HTTP 5xx up to 3 times.
	// Unknown failures are also treated as temporary.
	failureTemporary
	// failureVerification means the verification failed. Changing the access token doesn't change the result.
	failureVerification
	// failureNoAttestation means no attestation was found. Changing the access token doesn't change the result.
	failureNoAttestation
	// failureNotFound is HTTP 404. It's caused by the lack of the permission,
	// a missing repository, or missing attestations. They can't be distinguished.
	failureNotFound
	// failureInvalidToken is HTTP 401. The access token is invalid (e.g. revoked or expired).
	failureInvalidToken
	// failureRateLimit means the rate limit is exceeded.
	failureRateLimit
	// failurePermission is HTTP 403 other than the rate limit (e.g. SAML SSO enforcement).
	// It may depend on the repository, so the access token is still used for other packages.
	failurePermission
	// failureAuth means GitHub CLI requires the authentication.
	// GH_HOST makes GitHub CLI skip the authentication check, so this rarely happens.
	// The next credential is tried, and the verification fails if all credentials fail.
	failureAuth
)

func (f failure) String() string {
	// The order must be the same as the constants.
	names := [...]string{
		"none",
		"temporary",
		"verification",
		"no_attestation",
		"not_found",
		"invalid_token",
		"rate_limit",
		"permission",
		"auth",
	}
	if f < 0 || int(f) >= len(names) {
		return "unknown"
	}
	return names[f]
}

// e.g. expected SourceRepositoryURI to be https://github.com/foo/bar, got https://github.com/foo/baz
var certExtensionMismatch = regexp.MustCompile(`expected \w+ to be .*, got `)

// classify classifies the output of gh attestation verify.
func classify(out string) failure {
	switch {
	case strings.Contains(out, "no attestations found"):
		return failureNoAttestation
	case strings.Contains(out, "verifying with issuer"), certExtensionMismatch.MatchString(out):
		return failureVerification
	case strings.Contains(out, "gh auth login"), strings.Contains(out, "set the GH_TOKEN environment variable"):
		return failureAuth
	default:
		return classifyHTTP(out)
	}
}

// classifyHTTP classifies the HTTP error of gh attestation verify.
// The format is `HTTP <status>: <message> (<URL>)`.
func classifyHTTP(out string) failure {
	switch {
	case strings.Contains(out, "HTTP 401"), strings.Contains(out, "Bad credentials"):
		return failureInvalidToken
	case (strings.Contains(out, "HTTP 403") || strings.Contains(out, "HTTP 429")) && strings.Contains(strings.ToLower(out), "rate limit"):
		return failureRateLimit
	case strings.Contains(out, "HTTP 403"):
		return failurePermission
	case strings.Contains(out, "HTTP 404"):
		return failureNotFound
	default:
		return failureTemporary
	}
}
