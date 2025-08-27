package commands

import (
	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "noji",
		Short: "Noji - NO JIRA!!!1!!!",
		Long:  "Noji is a CLI wrapper around opencode for PRs and tickets.",
	}
	return rootCmd
}
