package minisign

import (
	"context"
	"sync"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/sirupsen/logrus"
)

type Verifier struct{}

func NewVerifier() *Verifier {
	return &Verifier{}
}

func (v *Verifier) Package() *config.Package {
	return &config.Package{
		Package: &aqua.Package{
			Name:    "jedisct1/minisign",
			Version: Version,
		},
		PackageInfo: &registry.PackageInfo{
			Type:                "github_release",
			RepoOwner:           "jedisct1",
			RepoName:            "minisign",
			Asset:               "minisign-{{.Version}}-{{.OS}}.{{.Format}}",
			Format:              "zip",
			Rosetta2:            true,
			WindowsARMEmulation: true,
			Replacements: map[string]string{
				"darwin":  "macos",
				"windows": "win64",
				"amd64":   "x86_64",
				"arm64":   "aarch64",
			},
			Overrides: []*registry.Override{
				{
					GOOS:   "linux",
					Format: "tar.gz",
					Files: []*registry.File{
						{
							Name: "minisign",
							Src:  "minisign-{{.OS}}/{{.Arch}}/minisign",
						},
					},
				},
				{
					GOOS: "windows",
					Files: []*registry.File{
						{
							Name: "minisign",
							Src:  "minisign-win64/minisign.exe",
						},
					},
				},
			},
			SupportedEnvs: []string{
				"darwin",
				"windows",
				"amd64",
			},
		},
	}
}

func (v *Verifier) Checksums() map[string]string {
	return Checksums()
}

func (v *Verifier) Enabled(pkg *registry.PackageInfo) bool {
	if pkg.Minisign == nil {
		return false
	}
	if pkg.Minisign.Enabled == nil {
		return true
	}
	return *pkg.Minisign.Enabled
}

func (v *Verifier) SupportedConfig() bool {
	return true
}

func (v *Verifier) Signature(ctx context.Context, logE *logrus.Entry) (*registry.DownloadedFile, string, error) {
	return v.Package().PackageInfo.Minisign.ToDownloadedFile(), "", nil
}

func (v *Verifier) Command(verifiedFilePath, sigPath string) (*sync.Mutex, int, []string) {
	return nil, 1, []string{
		"-Vm",
		verifiedFilePath,
		"-P",
		v.Package().PackageInfo.Minisign.PublicKey,
		"-x",
		sigPath,
	}
}
