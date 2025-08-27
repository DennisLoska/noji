package main

import (
	"fmt"
	"os"

	"github.com/dennis/noji/internal/commands"
)

func main() {
	rootCmd := commands.BuildRoot()
	if err := rootCmd.Execute(); err != nil {
		// Swallow the sentinel early-exit error used by -v/--version
		if err.Error() == "__noji_exit__" {
			// ensure no additional output
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
