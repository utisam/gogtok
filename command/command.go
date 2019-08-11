package command

import "github.com/spf13/cobra"

// New is a constructor of gogtok commands
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use: "gogtok",
	}

	cmd.AddCommand(newGlue())
	cmd.AddCommand(newImport())
	cmd.AddCommand(newList())
	cmd.AddCommand(newNew())
	cmd.AddCommand(newPackage())

	return cmd
}
