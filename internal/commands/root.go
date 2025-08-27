package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/dennis/noji/internal/commands/output"
	"github.com/dennis/noji/internal/config"
	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	var colorFlag string
	var editorFlag string
	var versionFlag bool

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

			// If --editor is provided, persist override for this process via env
			if strings.TrimSpace(editorFlag) != "" {
				os.Setenv("NOJI_EDITOR_OVERRIDE", editorFlag)
			}

			// Handle global version flag early and exit
			if versionFlag {
				// print short version like v0.1.0
				v := mustShortVersion()
				fmt.Fprintln(cmd.OutOrStdout(), v)
				// prevent command RunE from executing
				return cobra.ErrSubCommandRequired
			}

			// Optional: add hidden flag to show paths when verbose/debugging
			_ = cfg
			_ = prompts
			return nil
		},
	}

	// Ensure that running plain `noji` executes PersistentPreRunE (seeding config/prompts)
	// and then shows help to preserve UX.
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Print help and return nil to indicate success
		return cmd.Help()
	}

	rootCmd.PersistentFlags().StringVar(&colorFlag, "color", "auto", "color output: auto|always|never")
	rootCmd.PersistentFlags().StringVar(&editorFlag, "editor", "", "preferred editor binary or command (overrides config)")
	rootCmd.PersistentFlags().BoolVarP(&versionFlag, "version", "v", false, "print version and exit")
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
