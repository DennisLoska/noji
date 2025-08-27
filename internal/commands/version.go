package commands

import (
	"fmt"

	"github.com/dennis/noji/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	var short bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if short {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", version.Version)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Version: %s\n", version.Version)
			fmt.Fprintf(cmd.OutOrStdout(), "Commit:  %s\n", version.Commit)
			fmt.Fprintf(cmd.OutOrStdout(), "Date:    %s\n", version.Date)
			return nil
		},
	}
	cmd.Flags().BoolVar(&short, "short", false, "print only the version")
	return cmd
}
