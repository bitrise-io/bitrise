package progress

import "fmt"

// Printer defines the interface for printing to output.
// This abstraction allows injecting custom output implementations for testing
// or directing output to different destinations.
type Printer interface {
	// PrintWithoutNewline prints text without adding a newline.
	PrintWithoutNewline(text string)

	// Println prints a newline.
	Println()
}

// FmtPrinter is the default Printer implementation using fmt.
// It prints to stdout.
type FmtPrinter struct{}

// NewFmtPrinter creates a new FmtPrinter.
func NewFmtPrinter() *FmtPrinter {
	return &FmtPrinter{}
}

// PrintWithoutNewline prints text to stdout without a newline.
func (p *FmtPrinter) PrintWithoutNewline(text string) {
	fmt.Print(text)
}

// Println prints a newline to stdout.
func (p *FmtPrinter) Println() {
	fmt.Println()
}
