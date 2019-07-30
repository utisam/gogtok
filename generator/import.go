package generator

import (
	"io"
	"strings"
)

type ImportPackage struct {
	Name string
	Path string
}

func (pkg *ImportPackage) String() string {
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

func ParseImportPackage(s string) (*ImportPackage, error) {
	s = strings.TrimSpace(s)

	pair := strings.SplitN(s, " ", 2)
	if len(pair) == 1 {
		return &ImportPackage{
			Path: strings.Trim(s, `"`),
		}, nil
	}

	return &ImportPackage{
		Name: pair[0],
		Path: strings.Trim(pair[1], `"`),
	}, nil
}

func RenderImport(w io.Writer, pkgs []*ImportPackage) (err error) {
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
