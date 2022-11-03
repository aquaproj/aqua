package aqua

func (chk *Checksum) GetEnabled() bool {
	if chk == nil || chk.Enabled == nil {
		return false
	}
	return *chk.Enabled
}
