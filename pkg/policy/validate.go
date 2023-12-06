package policy

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Validator interface {
	Validate(p string) error
	Allow(p string) error
	Deny(p string) error
	Warn(logE *logrus.Entry, policyFilePath string, updated bool) error
}

type ValidatorImpl struct {
	rootDir  string
	fs       afero.Fs
	disabled bool
}

func NewValidator(param *config.Param, fs afero.Fs) *ValidatorImpl {
	return &ValidatorImpl{
		rootDir:  param.RootDir,
		fs:       fs,
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
	fs := v.fs
	if err := osfile.MkdirAll(fs, dir); err != nil {
		return fmt.Errorf("create a directory where the policy file is stored: %w", err)
	}
	f1, err := fs.Open(p)
	if err != nil {
		return fmt.Errorf("open a policy file: %w", err)
	}
	defer f1.Close()
	f2, err := fs.Create(policyPath)
	if err != nil {
		return fmt.Errorf("create a policy file: %w", err)
	}
	defer f2.Close()
	if _, err := io.Copy(f2, f1); err != nil {
		return fmt.Errorf("copy a policy file: %w", err)
	}
	warnFilePath := filepath.Join(v.rootDir, "policy-warnings", normalizedP)
	warnExist, err := afero.Exists(fs, warnFilePath)
	if err != nil {
		return fmt.Errorf("check if a warn file exists: %w", err)
	}
	if !warnExist {
		return nil
	}
	if err := fs.Remove(warnFilePath); err != nil {
		return fmt.Errorf("remove a warn file: %w", err)
	}
	return nil
}

func (v *ValidatorImpl) Deny(p string) error {
	normalizedP := normalizePath(p)
	policyPath := filepath.Join(v.rootDir, "policies", normalizedP)
	fs := v.fs

	// remove allow file
	policyExist, err := afero.Exists(fs, policyPath)
	if err != nil {
		return fmt.Errorf("check if a policy file exists: %w", err)
	}
	if policyExist {
		if err := fs.Remove(policyPath); err != nil {
			return fmt.Errorf("remove a policy file: %w", err)
		}
	}

	warnFilePath := filepath.Join(v.rootDir, "policy-warnings", normalizedP)
	warnFileDir := filepath.Dir(warnFilePath)
	if err := osfile.MkdirAll(fs, warnFileDir); err != nil {
		return fmt.Errorf("create a directory where the policy warning file is stored: %w", err)
	}
	warnFile, err := v.fs.Create(warnFilePath)
	if err != nil {
		return fmt.Errorf("create a policy warn file: %w", err)
	}
	defer warnFile.Close()
	return nil
}

func (v *ValidatorImpl) Warn(logE *logrus.Entry, policyFilePath string, updated bool) error {
	warnFilePath := filepath.Join(v.rootDir, "policy-warnings", normalizePath(policyFilePath))
	fs := v.fs
	f, err := afero.Exists(fs, warnFilePath)
	if err != nil {
		return fmt.Errorf("find a policy warning file: %w", err)
	}
	if f {
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
	logE.WithFields(logrus.Fields{
		"policy_file": policyFilePath,
		"doc":         "https://aquaproj.github.io/docs/reference/codes/003",
	}).Warnf(msg, policyFilePath, policyFilePath)
	return nil
}

func (v *ValidatorImpl) Validate(p string) error {
	if v.disabled {
		return nil
	}
	policyPath := filepath.Join(v.rootDir, "policies", normalizePath(p))
	f, err := afero.Exists(v.fs, policyPath)
	if err != nil {
		return fmt.Errorf("find a policy file: %w", err)
	}
	if !f {
		return errPolicyNotFound
	}
	b1, err := afero.ReadFile(v.fs, p)
	if err != nil {
		return fmt.Errorf("read a policy file: %w", err)
	}
	b2, err := afero.ReadFile(v.fs, policyPath)
	if err != nil {
		return fmt.Errorf("read a policy file: %w", err)
	}
	if string(b1) == string(b2) {
		return nil
	}
	return errPolicyUpdated
}
