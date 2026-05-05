package asset_test

const (
	osDarwin           = "darwin"
	osLinux            = "linux"
	osLinuxCapitalized = "Linux"
	osLinuxUpper       = "LINUX"
	osWindows          = "windows"

	archAmd64  = "amd64"
	archArm64  = "arm64"
	archX86_64 = "x86_64"

	formatZip   = "zip"
	formatTarGz = "tar.gz"
	formatRaw   = "raw"

	assetFOOTarGz             = "FOO.tar.gz"
	assetFOOLinuxAmd64TarGz   = "FOO_LINUX_AMD64.tar.gz"
	assetFooLinuxAmd64TarGz   = "foo_linux_amd64.tar.gz"
	assetTemplateOSArchFormat = "tool-{{.OS}}-{{.Arch}}.{{.Format}}"

	repoCliCli = "cli/cli"
	versionV1  = "v1.0.0"

	keyName = "name"
)
