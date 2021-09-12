package controller

func replace(key string, replacements map[string]string) string {
	a := replacements[key]
	if a == "" {
		return key
	}
	return a
}
