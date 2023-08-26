package ptr

func String(a string) *string {
	return &a
}

func Int(a int) *int {
	return &a
}

func Int64(a int64) *int64 {
	return &a
}

func Bool(a bool) *bool {
	return &a
}
