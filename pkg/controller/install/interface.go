package install

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}
