package list

type ConfigFinder interface {
	Find(wd, configFilePath string, globalConfigFilePaths ...string) (string, error)
	Finds(wd, configFilePath string) []string
}
