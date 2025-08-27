package commands

import "github.com/spf13/cobra"

// Expose ready-to-use commands
func BuildRoot() *cobra.Command {
	root := NewRoot()
	root.AddCommand(newModelsCmd())
	root.AddCommand(newUseCmd())
	root.AddCommand(newPRCmd())
	root.AddCommand(newTicketCmd())
	return root
}
