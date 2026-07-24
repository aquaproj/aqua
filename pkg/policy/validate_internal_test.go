package policy

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
)

// newTestValidator creates a validator whose root directory is empty, and a
// policy file outside of it. It returns the validator and the policy file path.
func newTestValidator(t *testing.T, policy string) (*ValidatorImpl, string) {
	t.Helper()
	policyFilePath := filepath.Join(t.TempDir(), "aqua-policy.yaml")
	if err := os.WriteFile(policyFilePath, []byte(policy), 0o600); err != nil {
		t.Fatal(err)
	}
	v := NewValidator(&config.Param{
		RootDir: t.TempDir(),
	})
	return v, policyFilePath
}

func exists(t *testing.T, p string) bool {
	t.Helper()
	if _, err := os.Stat(p); err == nil {
		return true
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	return false
}

func TestValidatorImpl_Allow(t *testing.T) {
	t.Parallel()

	v, policyFilePath := newTestValidator(t, "packages:\n")
	if err := v.Allow(policyFilePath); err != nil {
		t.Fatal(err)
	}
	// The policy file is copied into the root directory so that a later change
	// to it can be detected.
	b, err := os.ReadFile(filepath.Join(v.rootDir, "policies", normalizePath(policyFilePath)))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "packages:\n" {
		t.Fatalf("the copied policy file is %q, want %q", string(b), "packages:\n")
	}
}

func TestValidatorImpl_Allow_removeWarnFile(t *testing.T) {
	t.Parallel()

	v, policyFilePath := newTestValidator(t, "")
	// Deny creates the warning file, which Allow must remove. Otherwise the
	// policy file would keep being treated as denied.
	if err := v.Deny(policyFilePath); err != nil {
		t.Fatal(err)
	}
	if err := v.Allow(policyFilePath); err != nil {
		t.Fatal(err)
	}
	if warnFilePath := filepath.Join(v.rootDir, "policy-warnings", normalizePath(policyFilePath)); exists(t, warnFilePath) {
		t.Fatal("the warning file still exists")
	}
}

func TestValidatorImpl_Allow_policyFileNotFound(t *testing.T) {
	t.Parallel()

	v, policyFilePath := newTestValidator(t, "")
	if err := os.Remove(policyFilePath); err != nil {
		t.Fatal(err)
	}
	if err := v.Allow(policyFilePath); err == nil {
		t.Fatal("an error must occur")
	}
}

func TestValidatorImpl_Deny(t *testing.T) {
	t.Parallel()

	v, policyFilePath := newTestValidator(t, "")
	if err := v.Deny(policyFilePath); err != nil {
		t.Fatal(err)
	}
	if warnFilePath := filepath.Join(v.rootDir, "policy-warnings", normalizePath(policyFilePath)); !exists(t, warnFilePath) {
		t.Fatal("the warning file isn't created")
	}
}

func TestValidatorImpl_Deny_removeAllowedPolicyFile(t *testing.T) {
	t.Parallel()

	v, policyFilePath := newTestValidator(t, "")
	if err := v.Allow(policyFilePath); err != nil {
		t.Fatal(err)
	}
	if err := v.Deny(policyFilePath); err != nil {
		t.Fatal(err)
	}
	if err := v.Validate(policyFilePath); !errors.Is(err, errPolicyNotFound) {
		t.Fatalf("the policy file is still allowed: %v", err)
	}
}

func TestValidatorImpl_Warn(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)

	v, policyFilePath := newTestValidator(t, "")
	if err := v.Warn(logger, policyFilePath, false); err != nil {
		t.Fatal(err)
	}
	// Once the policy file is denied, the warning file exists and the warning
	// isn't repeated.
	if err := v.Deny(policyFilePath); err != nil {
		t.Fatal(err)
	}
	if err := v.Warn(logger, policyFilePath, true); err != nil {
		t.Fatal(err)
	}
}

func TestValidatorImpl_Validate(t *testing.T) {
	t.Parallel()

	v, policyFilePath := newTestValidator(t, "packages:\n")
	if err := v.Allow(policyFilePath); err != nil {
		t.Fatal(err)
	}
	if err := v.Validate(policyFilePath); err != nil {
		t.Fatal(err)
	}
}

func TestValidatorImpl_Validate_notFound(t *testing.T) {
	t.Parallel()

	v, policyFilePath := newTestValidator(t, "")
	if err := v.Validate(policyFilePath); !errors.Is(err, errPolicyNotFound) {
		t.Fatalf("wanted %v, got %v", errPolicyNotFound, err)
	}
}

func TestValidatorImpl_Validate_updated(t *testing.T) {
	t.Parallel()

	v, policyFilePath := newTestValidator(t, "packages:\n")
	if err := v.Allow(policyFilePath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(policyFilePath, []byte("packages:\n- name: foo\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := v.Validate(policyFilePath); !errors.Is(err, errPolicyUpdated) {
		t.Fatalf("wanted %v, got %v", errPolicyUpdated, err)
	}
}

// A policy file that exists in the root directory but can't be stat'd must not
// be reported as errPolicyNotFound: a permission problem would then be
// misdiagnosed, and the user told to allow a policy that is already there.
func TestValidatorImpl_Validate_unreadable(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}

	v, policyFilePath := newTestValidator(t, "packages:\n")
	if err := v.Allow(policyFilePath); err != nil {
		t.Fatal(err)
	}
	// Make the directory that holds the copied policy file unsearchable, so
	// stat of the policy file fails with a permission error rather than
	// os.ErrNotExist.
	policyDir := filepath.Dir(filepath.Join(v.rootDir, "policies", normalizePath(policyFilePath)))
	if err := os.Chmod(policyDir, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(policyDir, 0o755)
	})
	err := v.Validate(policyFilePath)
	if err == nil {
		t.Fatal("an error must be returned when the policy file can't be stat'd")
	}
	if errors.Is(err, errPolicyNotFound) {
		t.Fatal("a permission error must not be reported as errPolicyNotFound")
	}
}

func TestValidatorImpl_Validate_disabled(t *testing.T) {
	t.Parallel()

	v := NewValidator(&config.Param{
		RootDir:       t.TempDir(),
		DisablePolicy: true,
	})
	// The policy file doesn't exist, but the validation is skipped.
	if err := v.Validate(filepath.Join(t.TempDir(), "aqua-policy.yaml")); err != nil {
		t.Fatal(err)
	}
}
