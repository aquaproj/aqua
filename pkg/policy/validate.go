package policy

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
)

type Validator interface {
	Validate(p string) error
	Allow(p string) error
	Deny(p string) error
	Warn(logger *slog.Logger, policyFilePath string, updated bool) error
}

type ValidatorImpl struct {
	rootDir  string
	disabled bool
}

func NewValidator(param *config.Param) *ValidatorImpl {
	return &ValidatorImpl{
		rootDir:  param.RootDir,
		disabled: param.DisablePolicy,
	}
}

var (
	errPolicyNotFound = errors.New("the policy file isn't found")
	errPolicyUpdated  = errors.New("the policy file updated")
)

func normalizePath(p string) string {
	// Replace ":" to "_" on Windows.
	// https://www.ibm.com/docs/en/spectrum-archive-sde/2.4.1.0?topic=tips-file-name-characters
	// On Windows systems, files and directory names cannot be created with a colon (:). But if a file or directory name is created with a colon on a Linux or Mac operating system, then moved to a Windows system, percent encoding is used to include the colon in the name in the index.
	// Note: On a Windows system, the : (colon) character appears as “%3A” in the index with percent encoding. The colon appears as an “_” (underscore) in the file name on Windows Explorer, or in the command prompt.
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(p, ":", "_")
	}
	return p
}

func (v *ValidatorImpl) Allow(p string) error {
	normalizedP := normalizePath(p)
	policyPath := filepath.Join(v.rootDir, "policies", normalizedP)
	dir := filepath.Dir(policyPath)
	if err := osfile.MkdirAll(dir); err != nil {
		return fmt.Errorf("create a directory where the policy file is stored: %w", err)
	}
	f1, err := os.Open(p)
	if err != nil {
		return fmt.Errorf("open a policy file: %w", err)
	}
	defer f1.Close()
	f2, err := os.Create(policyPath)
	if err != nil {
		return fmt.Errorf("create a policy file: %w", err)
	}
	defer f2.Close()
	if _, err := io.Copy(f2, f1); err != nil {
		return fmt.Errorf("copy a policy file: %w", err)
	}
	warnFilePath := filepath.Join(v.rootDir, "policy-warnings", normalizedP)
	if !osfile.Exists(warnFilePath) {
		return nil
	}
	if err := os.Remove(warnFilePath); err != nil {
		return fmt.Errorf("remove a warn file: %w", err)
	}
	return nil
}

func (v *ValidatorImpl) Deny(p string) error {
	normalizedP := normalizePath(p)
	policyPath := filepath.Join(v.rootDir, "policies", normalizedP)

	// remove allow file
	if osfile.Exists(policyPath) {
		if err := os.Remove(policyPath); err != nil {
			return fmt.Errorf("remove a policy file: %w", err)
		}
	}

	warnFilePath := filepath.Join(v.rootDir, "policy-warnings", normalizedP)
	warnFileDir := filepath.Dir(warnFilePath)
	if err := osfile.MkdirAll(warnFileDir); err != nil {
		return fmt.Errorf("create a directory where the policy warning file is stored: %w", err)
	}
	warnFile, err := os.Create(warnFilePath)
	if err != nil {
		return fmt.Errorf("create a policy warn file: %w", err)
	}
	defer warnFile.Close()
	return nil
}

func (v *ValidatorImpl) Warn(logger *slog.Logger, policyFilePath string, updated bool) error {
	warnFilePath := filepath.Join(v.rootDir, "policy-warnings", normalizePath(policyFilePath))
	if osfile.Exists(warnFilePath) {
		return nil
	}
	msg := `The policy file is ignored unless it is allowed by "aqua policy allow" command.

$ aqua policy allow "%s"

If you want to keep ignoring the policy file without the warning, please run "aqua policy deny" command.

$ aqua policy deny "%s"

 `
	if updated {
		msg = `The policy file is changed. ` + msg
	}
	logger.Warn(fmt.Sprintf(msg, policyFilePath, policyFilePath),
		"policy_file", policyFilePath,
		"doc", "https://aquaproj.github.io/docs/reference/codes/003")
	return nil
}

func (v *ValidatorImpl) Validate(p string) error {
	if v.disabled {
		return nil
	}
	policyPath := filepath.Join(v.rootDir, "policies", normalizePath(p))
	if !osfile.Exists(policyPath) {
		return errPolicyNotFound
	}
	b1, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("read a policy file: %w", err)
	}
	b2, err := os.ReadFile(policyPath)
	if err != nil {
		return fmt.Errorf("read a policy file: %w", err)
	}
	if string(b1) == string(b2) {
		return nil
	}
	return errPolicyUpdated
}
