package asset_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/asset"
)

func TestOS(t *testing.T) {
	t.Parallel()

	os := &asset.OS{
		Name: "linux",
		OS:   "linux",
	}

	if os.Name != "linux" {
		t.Errorf("Expected Name to be 'linux', got %s", os.Name)
	}
	if os.OS != "linux" {
		t.Errorf("Expected OS to be 'linux', got %s", os.OS)
	}
}

func TestArch(t *testing.T) {
	t.Parallel()

	arch := &asset.Arch{
		Name: "amd64",
		Arch: "amd64",
	}

	if arch.Name != "amd64" {
		t.Errorf("Expected Name to be 'amd64', got %s", arch.Name)
	}
	if arch.Arch != "amd64" {
		t.Errorf("Expected Arch to be 'amd64', got %s", arch.Arch)
	}
}

func TestAssetInfo(t *testing.T) {
	t.Parallel()

	completeWindowsExt := true
	assetInfo := &asset.AssetInfo{
		Template:           "test-{{.Version}}-{{.OS}}-{{.Arch}}.{{.Format}}",
		OS:                 "linux",
		Arch:               "amd64",
		DarwinAll:          false,
		Format:             "tar.gz",
		Replacements:       map[string]string{"linux": "Linux"},
		Score:              0,
		CompleteWindowsExt: &completeWindowsExt,
	}

	if assetInfo.Template != "test-{{.Version}}-{{.OS}}-{{.Arch}}.{{.Format}}" {
		t.Errorf("Expected Template to match, got %s", assetInfo.Template)
	}
	if assetInfo.OS != "linux" {
		t.Errorf("Expected OS to be 'linux', got %s", assetInfo.OS)
	}
	if assetInfo.Arch != "amd64" {
		t.Errorf("Expected Arch to be 'amd64', got %s", assetInfo.Arch)
	}
	if assetInfo.DarwinAll != false {
		t.Errorf("Expected DarwinAll to be false, got %v", assetInfo.DarwinAll)
	}
	if assetInfo.Format != "tar.gz" {
		t.Errorf("Expected Format to be 'tar.gz', got %s", assetInfo.Format)
	}
	if assetInfo.Replacements["linux"] != "Linux" {
		t.Errorf("Expected Replacements['linux'] to be 'Linux', got %s", assetInfo.Replacements["linux"])
	}
	if assetInfo.Score != 0 {
		t.Errorf("Expected Score to be 0, got %d", assetInfo.Score)
	}
	if assetInfo.CompleteWindowsExt == nil || *assetInfo.CompleteWindowsExt != true {
		t.Errorf("Expected CompleteWindowsExt to be true, got %v", assetInfo.CompleteWindowsExt)
	}
}
