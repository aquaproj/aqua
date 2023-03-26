package policy

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Validator interface {
	Validate(p string) error
	Allow(p string) error
	Disallow(p string) error
	Warn(logE *logrus.Entry, policyFilePath string) error
}

type MockValidator struct {
	Err error
}

func (validator *MockValidator) Allow(p string) error {
	return validator.Err
}

func (validator *MockValidator) Disallow(p string) error {
	return validator.Err
}

func (validator *MockValidator) Validate(p string) error {
	return validator.Err
}

func (validator *MockValidator) Warn(logE *logrus.Entry, policyFilePath string) error {
	return validator.Err
}

type ValidatorImpl struct {
	rootDir string
	fs      afero.Fs
}

func NewValidator(param *config.Param, fs afero.Fs) *ValidatorImpl {
	return &ValidatorImpl{
		rootDir: param.RootDir,
		fs:      fs,
	}
}

var (
	errPolicyNotFound = errors.New("the policy file isn't found")
	errPolicyUpdated  = errors.New("the policy file updated")
)

func (validator *ValidatorImpl) Allow(p string) error {
	policyPath := filepath.Join(validator.rootDir, "policies", p)
	dir := filepath.Dir(policyPath)
	fs := validator.fs
	if err := util.MkdirAll(fs, dir); err != nil {
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
	warnFilePath := filepath.Join(validator.rootDir, "policy-warnings", p)
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

func (validator *ValidatorImpl) Disallow(p string) error {
	policyPath := filepath.Join(validator.rootDir, "policies", p)
	fs := validator.fs

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

	warnFilePath := filepath.Join(validator.rootDir, "policy-warnings", p)
	warnFileDir := filepath.Dir(warnFilePath)
	if err := util.MkdirAll(fs, warnFileDir); err != nil {
		return fmt.Errorf("create a directory where the policy warning file is stored: %w", err)
	}
	warnFile, err := validator.fs.Create(warnFilePath)
	if err != nil {
		return fmt.Errorf("create a policy warn file: %w", err)
	}
	defer warnFile.Close()
	return nil
}

func (validator *ValidatorImpl) Warn(logE *logrus.Entry, policyFilePath string) error {
	warnFilePath := filepath.Join(validator.rootDir, "policy-warnings", policyFilePath)
	fs := validator.fs
	f, err := afero.Exists(fs, warnFilePath)
	if err != nil {
		return fmt.Errorf("find a policy warning file: %w", err)
	}
	if f {
		return nil
	}
	logE.WithFields(logrus.Fields{
		"policy_file": policyFilePath,
	}).Warnf(`The policy file is ignored unless it is allowed by aqua allow-policy command.

$ aqua allow-policy "%s"

If you want to keep ignoring the policy file without the warning, please run disallow-policy command.

$ aqua disallow-policy "%s"

 `, policyFilePath, policyFilePath)
	return nil
}

func (validator *ValidatorImpl) Validate(p string) error {
	policyPath := filepath.Join(validator.rootDir, "policies", p)
	f, err := afero.Exists(validator.fs, policyPath)
	if err != nil {
		return fmt.Errorf("find a policy file: %w", err)
	}
	if !f {
		return errPolicyNotFound
	}
	b1, err := afero.ReadFile(validator.fs, p)
	if err != nil {
		return fmt.Errorf("read a policy file: %w", err)
	}
	b2, err := afero.ReadFile(validator.fs, policyPath)
	if err != nil {
		return fmt.Errorf("read a policy file: %w", err)
	}
	if string(b1) == string(b2) {
		return nil
	}
	return errPolicyUpdated
}
