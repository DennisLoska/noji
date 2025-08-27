package commands

import (
	"github.com/dennis/noji/internal/config"
	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "noji",
		Short: "Noji - NO JIRA!!!1!!!",
		Long:  "Noji is a CLI wrapper around opencode for PRs and tickets.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, prompts, err := config.EnsureConfig()
			if err != nil {
				return err
			}
			// Optional: add hidden flag to show paths when verbose/debugging
			_ = cfg
			_ = prompts
			return nil
		},
	}
	return rootCmd
}
