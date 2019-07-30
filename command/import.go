package command

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/utisam/gogtok/generator"
)

func newImport() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Generate an import statement",
		RunE: func(_ *cobra.Command, args []string) error {

			pkgs := make([]*generator.ImportPackage, len(args))
			for i, s := range args {
				pkg, err := generator.ParseImportPackage(s)
				if err != nil {
					return err
				}

				pkgs[i] = pkg
			}

			return generator.RenderImport(os.Stdout, pkgs)
		},
	}

	return cmd
}
