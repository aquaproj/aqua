package which

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}
