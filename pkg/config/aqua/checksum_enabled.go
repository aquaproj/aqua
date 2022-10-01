package aqua

import "os"

func (chk *Checksum) GetEnabled() bool {
	if os.Getenv("AQUA_EXPERIMENTAL_CHECKSUM_VERIFICATION") != "true" {
		return false
	}
	if chk == nil || chk.Enabled == nil {
		return false
	}
	return *chk.Enabled
}
