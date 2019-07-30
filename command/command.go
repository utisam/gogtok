package command

import "github.com/spf13/cobra"

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use: "gogtok",
	}

	cmd.AddCommand(newImport())
	cmd.AddCommand(newGlue())
	cmd.AddCommand(newNew())

	return cmd
}
