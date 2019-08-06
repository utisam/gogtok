package command

import (
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type importPackage struct {
	Name string
	Path string
}

func (pkg *importPackage) String() string {
	b := strings.Builder{}
	if pkg.Name != "" {
		b.WriteString(pkg.Name)
		b.WriteRune(' ')
	}
	b.WriteRune('"')
	b.WriteString(pkg.Path)
	b.WriteRune('"')

	return b.String()
}

func parseImportPackage(s string) (*importPackage, error) {
	s = strings.TrimSpace(s)

	pair := strings.SplitN(s, " ", 2)
	if len(pair) == 1 {
		return &importPackage{
			Path: strings.Trim(s, `"`),
		}, nil
	}

	return &importPackage{
		Name: pair[0],
		Path: strings.Trim(pair[1], `"`),
	}, nil
}

func renderImport(w io.Writer, pkgs []*importPackage) (err error) {
	switch len(pkgs) {
	case 0:
		return nil
	case 1:
		io.WriteString(w, "import ")
		io.WriteString(w, pkgs[0].String())
		io.WriteString(w, "\n")
		return nil
	}

	io.WriteString(w, "import (\n")
	for _, pkg := range pkgs {
		io.WriteString(w, "\t")
		io.WriteString(w, pkg.String())
		io.WriteString(w, "\n")
	}
	io.WriteString(w, ")\n")
	return nil
}

func newImport() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Generate an import statement",
		RunE: func(_ *cobra.Command, args []string) error {

			pkgs := make([]*importPackage, len(args))
			for i, s := range args {
				pkg, err := parseImportPackage(s)
				if err != nil {
					return err
				}

				pkgs[i] = pkg
			}

			return renderImport(os.Stdout, pkgs)
		},
	}

	return cmd
}
