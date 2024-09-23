package ghattestation

const Version = "v2.57.0"

func Checksums() map[string]string {
	return map[string]string{
		"darwin/amd64":  "7fc0310baa48b0de034ffb632022f4633547d4bebf3ce7b42ea3e77da22ca390",
		"darwin/arm64":  "58f90d614f557b81ff4ee09a747ab68d13793a2f306d4e0102ad8beb9ae490fb",
		"linux/amd64":   "d6b3621aa0ca383866716fc664d827a21bd1ac4a918a10c047121d8031892bf8",
		"linux/arm64":   "a85069b7469846ee4afb5ca758aa1d8d3f801039598d521737b0caa407eeea36",
		"windows/amd64": "54bda7bc0ea27feb495a0f465493d9422be9e145aabeb72ab11d65d93c1ceba6",
		"windows/arm64": "7c16069d838eca957a4252ba5bbfc7c0eeebb06dd4d99cfb967301d7a197e1b2",
	}
}
