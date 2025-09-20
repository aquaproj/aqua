package config

const (
	// formatRaw is the raw format type for packages without compression or archiving
	formatRaw = "raw"
)

// DefaultVerCnt is the default value for --limit/-l flag in command generate, update.
// It limits the number of versions to process or display in various operations.
const DefaultVerCnt int = 30
