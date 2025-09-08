package updatepolicy

import (
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Controller struct {
	fs           afero.Fs
	configFinder ConfigFinder
	policyReader      PolicyReader
}

func New(fs afero.Fs) *Controller {
	return &Controller{
		fs: fs,
	}
}

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

type PolicyReader interface {
	Read(policyFilePaths []string) ([]*policy.Config, error)
	Append(logE *logrus.Entry, aquaYAMLPath string, policies []*policy.Config, globalPolicyPaths map[string]struct{}) ([]*policy.Config, error)
}
