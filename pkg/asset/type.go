package asset

type OS struct {
	Name string
	OS   string
}

type Arch struct {
	Name string
	Arch string
}

type AssetInfo struct { //nolint:revive
	Template           string
	OS                 string
	Arch               string
	DarwinAll          bool
	Format             string
	Replacements       map[string]string
	Score              int
	CompleteWindowsExt *bool
}
