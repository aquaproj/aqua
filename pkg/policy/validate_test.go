package policy_test

import (
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/spf13/afero"
)

func TestValidator_Allow(t *testing.T) {
	t.Parallel()
	data := []struct {
		name           string
		rootDir        string
		configFilePath string
		files          map[string]string
		isErr          bool
	}{
		{
			name:           caseNormal,
			rootDir:        pathHomeFooLocalShare,
			configFilePath: pathHomeFooWorkspacePolicy,
			files: map[string]string{
				pathHomeFooWorkspacePolicy: "",
			},
		},
		{
			name:           "warn file exists",
			rootDir:        pathHomeFooLocalShare,
			configFilePath: pathHomeFooWorkspacePolicy,
			files: map[string]string{
				pathHomeFooWorkspacePolicy: "",
				"/home/foo/.local/share/aquaproj-aqua/policy-warnings/home/foo/workspace/aqua-policy.yaml": "",
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for name, body := range d.files {
				if err := afero.WriteFile(fs, name, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			validator := policy.NewValidator(&config.Param{
				RootDir: d.rootDir,
			}, fs)
			if err := validator.Allow(d.configFilePath); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}

func TestValidator_Deny(t *testing.T) {
	t.Parallel()
	data := []struct {
		name           string
		rootDir        string
		configFilePath string
		files          map[string]string
		isErr          bool
	}{
		{
			name:           caseNormal,
			rootDir:        pathHomeFooLocalShare,
			configFilePath: pathHomeFooWorkspacePolicy,
			files: map[string]string{
				pathHomeFooWorkspacePolicy: "",
			},
		},
		{
			name:           "remove allowed policy file",
			configFilePath: pathHomeFooWorkspacePolicy,
			rootDir:        pathHomeFooLocalShare,
			files: map[string]string{
				pathHomeFooWorkspacePolicy: "",
				pathPolicyApplied:          "",
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for name, body := range d.files {
				if err := afero.WriteFile(fs, name, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			validator := policy.NewValidator(&config.Param{
				RootDir: d.rootDir,
			}, fs)
			if err := validator.Deny(d.configFilePath); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}

func TestValidator_Warn(t *testing.T) {
	t.Parallel()
	data := []struct {
		name           string
		rootDir        string
		configFilePath string
		files          map[string]string
		isErr          bool
	}{
		{
			name:           caseNormal,
			rootDir:        pathHomeFooLocalShare,
			configFilePath: pathHomeFooWorkspacePolicy,
			files: map[string]string{
				pathHomeFooWorkspacePolicy: "",
			},
		},
		{
			name:           "warn file exists",
			configFilePath: pathHomeFooWorkspacePolicy,
			rootDir:        pathHomeFooLocalShare,
			files: map[string]string{
				pathHomeFooWorkspacePolicy: "",
				"/home/foo/.local/share/aquaproj-aqua/policy-warnings/home/foo/workspace/aqua-policy.yaml": "",
			},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for name, body := range d.files {
				if err := afero.WriteFile(fs, name, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			validator := policy.NewValidator(&config.Param{
				RootDir: d.rootDir,
			}, fs)
			if err := validator.Warn(logger, d.configFilePath, false); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}

func TestValidator_Validate(t *testing.T) {
	t.Parallel()
	data := []struct {
		name           string
		rootDir        string
		configFilePath string
		files          map[string]string
		isErr          bool
	}{
		{
			name:           caseNormal,
			rootDir:        pathHomeFooLocalShare,
			configFilePath: pathHomeFooWorkspacePolicy,
			files: map[string]string{
				pathHomeFooWorkspacePolicy: "",
				pathPolicyApplied:          "",
			},
		},
		{
			name:           "policy is not found",
			configFilePath: pathHomeFooWorkspacePolicy,
			rootDir:        pathHomeFooLocalShare,
			files: map[string]string{
				pathHomeFooWorkspacePolicy: "",
			},
			isErr: true,
		},
		{
			name:           "policy is changed",
			configFilePath: pathHomeFooWorkspacePolicy,
			rootDir:        pathHomeFooLocalShare,
			files: map[string]string{
				pathHomeFooWorkspacePolicy: "",
				pathPolicyApplied:          "a",
			},
			isErr: true,
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for name, body := range d.files {
				if err := afero.WriteFile(fs, name, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			validator := policy.NewValidator(&config.Param{
				RootDir: d.rootDir,
			}, fs)
			if err := validator.Validate(d.configFilePath); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}
