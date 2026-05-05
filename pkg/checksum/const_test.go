package checksum_test

const (
	algoSHA256 = "sha256"
	algoSHA512 = "sha512"

	osDarwin     = "darwin"
	osLinux      = "linux"
	osDarwinAmd  = "darwin/amd64"
	osDarwinArm  = "darwin/arm64"
	osLinuxAmd   = "linux/amd64"
	osLinuxArm   = "linux/arm64"
	osWindowsAmd = "windows/amd64"
	archAmd64    = "amd64"
	archArm64    = "arm64"

	assetNova320DarwinArm64 = "nova_3.2.0_darwin_arm64.tar.gz"

	pkgFoo               = "foo"
	pkgVersionV1         = "v1.0.0"
	pkgVersion100        = "1.0.0"
	pkgTypeGitHubRelease = "github_release"

	helloWorld    = "hello world"
	checksumValue = "89f744a88dad0e73866d06e79afccd5476152770c70101361566b234b0722101"

	fileAquaYaml         = "aqua.yaml"
	fileAquaChecksums    = "aqua-checksums.json"
	fileDotAquaChecksums = ".aqua-checksums.json"

	pathHomeFooAquaYaml         = "/home/foo/aqua.yaml"
	pathHomeFooAquaChecksums    = "/home/foo/aqua-checksums.json"
	pathHomeFooDotAquaChecksums = "/home/foo/.aqua-checksums.json"
)
