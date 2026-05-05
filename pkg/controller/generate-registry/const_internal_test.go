package genrgst

const (
	pkgFoo     = "foo"
	caseNormal = "normal"

	repoOwner             = "owner"
	repoName              = "repo"
	fileChecksumsTxt      = "checksums.txt"
	fileChecksumsTxtSig   = "checksums.txt.sig"
	fileChecksumsKeyless  = "checksums.txt-keyless.sig"
	fileChecksumsKeylessP = "checksums.txt-keyless.pem"
	regexpCertIdentity    = `^https://github\.com/owner/repo/\.github/workflows/.+\.ya?ml@refs/tags/\Q{{.Version}}\E$`
)
