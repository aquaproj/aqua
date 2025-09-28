package aqua

// GetEnabled returns whether checksum validation is enabled.
// It returns false if the Checksum struct or Enabled field is nil.
func (c *Checksum) GetEnabled() bool {
	if c == nil || c.Enabled == nil {
		return false
	}
	return *c.Enabled
}
