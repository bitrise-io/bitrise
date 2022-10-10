package corelog

// Producer ...
type Producer string

const (
	// BitriseCLI ...
	BitriseCLI Producer = "bitrise_cli"
	// Step ...
	Step Producer = "step"
)

// Level ...
type Level string

const (
	// ErrorLevel ...
	ErrorLevel Level = "error"
	// WarnLevel ...
	WarnLevel Level = "warn"
	// InfoLevel ...
	InfoLevel Level = "info"
	// DoneLevel ...
	DoneLevel Level = "done"
	// NormalLevel ...
	NormalLevel Level = "normal"
	// DebugLevel ...
	DebugLevel Level = "debug"
)

type ANSIColorCode string

const (
	RedCode     ANSIColorCode = "\x1b[31;1m"
	YellowCode  ANSIColorCode = "\x1b[33;1m"
	BlueCode    ANSIColorCode = "\x1b[34;1m"
	GreenCode   ANSIColorCode = "\x1b[32;1m"
	MagentaCode ANSIColorCode = "\x1b[35;1m"
	ResetCode   ANSIColorCode = "\x1b[0m"
)
