package config

const (
	// formatRaw is the raw format type for packages without compression or archiving
	formatRaw = "raw"

	tmplVersion = "Version"
	tmplSemVer  = "SemVer"
	tmplGOOS    = "GOOS"
	tmplGOARCH  = "GOARCH"
	tmplArch    = "Arch"
	tmplFormat  = "Format"
	tmplVars    = "Vars"

	pkgNameAqua = "aqua"
	osDarwin    = "darwin"
	osWindows   = "windows"
)

// DefaultVerCnt is the default value for --limit/-l flag in command generate, update.
// It limits the number of versions to process or display in various operations.
const DefaultVerCnt int = 30

const (
	archAmd64 = "amd64"
	formatZip = "zip"
)

const repoOwnerAquaproj = "aquaproj"

const (
	versionV077  = "v0.7.7"
	repoOwnerCli = "cli"
	tmplGhAsset  = "gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}"
	fileBinGhExe = "bin/gh.exe"
)
