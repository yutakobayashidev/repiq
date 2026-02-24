package cli

import (
	"fmt"
	"io"
)

// Run executes the CLI with the given arguments.
func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: repiq <scheme>:<identifier> [...]")
		return fmt.Errorf("no targets specified")
	}
	return nil
}
