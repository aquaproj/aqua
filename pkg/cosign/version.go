package cosign

const Version = "v2.0.0"

func Checksums() map[string]string {
	return map[string]string{
		"darwin/amd64":  "d2c8fc0edb42a1e9745da1c43a2928cee044f3b8a1b8df64088a384c7e6f5b5d",
		"darwin/arm64":  "9d7821e1c05da4b07513729cb00d1070c9a95332c66d90fa593ed77d8c72ca2a",
		"linux/amd64":   "169a53594c437d53ffc401b911b7e70d453f5a2c1f96eb2a736f34f6356c4f2b",
		"linux/arm64":   "8132cb2fb99a4c60ba8e03b079e12462c27073028a5d08c07ecda67284e0c88d",
		"windows/amd64": "e78e7464dc0eda1d6ec063ac2738f4d1418b19dd19f999aa37e1679d5d3af82e",
	}
}
