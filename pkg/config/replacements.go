package config

type ReplacementsOverride struct {
	GOOS         string
	GOArch       string
	Replacements map[string]string
}

func replace(key string, replacements map[string]string) string {
	a := replacements[key]
	if a == "" {
		return key
	}
	return a
}
