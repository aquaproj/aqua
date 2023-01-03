package cosign

const Version = "v1.13.1"

func Checksums() map[string]string {
	return map[string]string{
		"darwin/amd64":  "1d164b8b1fcfef1e1870d809edbb9862afd5995cab63687a440b84cca5680ecf",
		"darwin/arm64":  "02bef878916be048fd7dcf742105639f53706a59b5b03f4e4eaccc01d05bc7ab",
		"linux/amd64":   "a50651a67b42714d6f1a66eb6773bf214dacae321f04323c0885f6a433051f95",
		"linux/arm64":   "a7a79a52c7747e2c21554cad4600e6c7130c0429017dd258f9c558d957fa9090",
		"windows/amd64": "78a2774b68b995cc698944f6c235b1c93dcb6d57593a58a565ee7a56d64e4b85",
	}
}
