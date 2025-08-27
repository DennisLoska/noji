package main

import (
	"fmt"
	"os"

	"github.com/dennis/noji/internal/commands"
)

func main() {
	rootCmd := commands.BuildRoot()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
