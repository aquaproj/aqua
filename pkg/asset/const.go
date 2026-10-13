package asset

const (
	osDarwin  = "darwin"
	osLinux   = "linux"
	osWindows = "windows"

	// Capitalized OS names used as replacement values
	osLinuxCapitalized = "Linux"

	archAmd64  = "amd64"
	archArm64  = "arm64"
	archX86_64 = "x86_64"
	// archArm is "arm" as release names write it, which the parser reads as arm64.
	archArm = "arm"
)
