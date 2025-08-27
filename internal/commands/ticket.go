package commands

import (
	"github.com/spf13/cobra"
)

func newTicketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ticket",
		Short: "Work with tickets",
	}
	cmd.AddCommand(newTicketUpdateCmd())
	return cmd
}

func newTicketUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update a ticket using opencode",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrompt("ticket_update.txt")
		},
	}
}
