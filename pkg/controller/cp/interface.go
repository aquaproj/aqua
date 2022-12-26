package cp

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}
