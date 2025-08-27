package commands

import (
	"fmt"
	"strings"

	"github.com/dennis/noji/internal/commands/output"
	"github.com/dennis/noji/internal/config"
	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	var colorFlag string

	rootCmd := &cobra.Command{
		Use:   "noji",
		Short: "Noji - NO JIRA!!!1!!!",
		Long:  "Noji is a CLI wrapper around opencode for PRs and tickets.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, prompts, err := config.EnsureConfig()
			if err != nil {
				return err
			}
			// initialize color output
			mode, err := outputParseMode(colorFlag)
			if err != nil {
				return err
			}
			output.Init(mode)

			// Optional: add hidden flag to show paths when verbose/debugging
			_ = cfg
			_ = prompts
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&colorFlag, "color", "auto", "color output: auto|always|never")
	rootCmd.PersistentFlags().Lookup("color").NoOptDefVal = "auto"
	rootCmd.RegisterFlagCompletionFunc("color", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"auto", "always", "never"}, cobra.ShellCompDirectiveNoFileComp
	})

	return rootCmd
}

// small shim to keep package import clean in root
func outputParseMode(s string) (output.Mode, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "auto", "":
		return output.ModeAuto, nil
	case "always":
		return output.ModeAlways, nil
	case "never":
		return output.ModeNever, nil
	default:
		return output.ModeAuto, fmt.Errorf("invalid color mode: %s (want auto|always|never)", s)
	}
}
