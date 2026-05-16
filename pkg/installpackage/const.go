package installpackage

const (
	pkgNameAqua          = "aqua"
	pkgTypeGitHubRelease = "github_release"
	algoSHA256           = "sha256"
	algoTypeRegexp       = "regexp"
	regexpSHA256         = `^(\b[A-Fa-f0-9]{64}\b)`
	regexpSHA256WithFile = `^\b[A-Fa-f0-9]{64}\b\s+(\S+)$`
)

const osDarwin = "darwin"
