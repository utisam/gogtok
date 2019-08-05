package command

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

func newList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List declarations of the source file",
	}

	cmd.AddCommand(newListFuncs())
	cmd.AddCommand(newListValues())
	cmd.AddCommand(newListTypes())
	cmd.AddCommand(newListFields())

	return cmd
}

func inspectFiles(filenames []string, fn func(f *ast.File) error) error {
	for _, filename := range filenames {
		reader, err := os.Open(filename)
		if err != nil {
			return err
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, filename, reader, parser.Mode(0))
		if err != nil {
			return err
		}

		if err := fn(f); err != nil {
			return err
		}
	}
	return nil
}

func compilePattern(p string) (*regexp.Regexp, error) {
	if p == "" {
		return nil, nil
	}
	return regexp.Compile(p)
}

func matchPattern(p *regexp.Regexp, s string) bool {
	return p == nil || p.MatchString(s)
}

func newListFuncs() *cobra.Command {
	var pattern string

	cmd := &cobra.Command{
		Use:   "funcs",
		Short: "List functions of the file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			patternRegexp, err := compilePattern(pattern)
			if err != nil {
				return err
			}

			inspectFiles(args, func(f *ast.File) error {
				for _, decl := range f.Decls {
					fnDecl, ok := decl.(*ast.FuncDecl)
					if !ok {
						continue
					}

					name := fnDecl.Name.Name
					if matchPattern(patternRegexp, name) {
						fmt.Println(name)
					}
				}
				return nil
			})
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&pattern, "pattern", "p", pattern, "Only print names matching the pattern")

	return cmd
}

func newListValues() *cobra.Command {
	var pattern string

	cmd := &cobra.Command{
		Use:   "values",
		Short: "List variables of the file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			patternRegexp, err := compilePattern(pattern)
			if err != nil {
				return err
			}

			inspectFiles(args, func(f *ast.File) error {
				for _, decl := range f.Decls {
					genDecl, ok := decl.(*ast.GenDecl)
					if !ok || (genDecl.Tok != token.VAR && genDecl.Tok != token.CONST) {
						continue
					}

					for _, spec := range genDecl.Specs {
						valSpec, ok := spec.(*ast.ValueSpec)
						if !ok {
							continue
						}

						for _, nameIdent := range valSpec.Names {
							name := nameIdent.Name
							if matchPattern(patternRegexp, name) {
								fmt.Println(name)
							}
						}
					}
				}
				return nil
			})
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&pattern, "pattern", "p", pattern, "Only print names matching the pattern")

	return cmd
}

func newListTypes() *cobra.Command {
	var pattern string

	cmd := &cobra.Command{
		Use:   "types",
		Short: "List types of the file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			patternRegexp, err := compilePattern(pattern)
			if err != nil {
				return err
			}

			inspectFiles(args, func(f *ast.File) error {
				for _, decl := range f.Decls {
					genDecl, ok := decl.(*ast.GenDecl)
					if !ok || genDecl.Tok != token.TYPE {
						continue
					}

					for _, spec := range genDecl.Specs {
						typeSpec, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}

						name := typeSpec.Name.Name
						if matchPattern(patternRegexp, name) {
							fmt.Println(name)
						}
					}
				}
				return nil
			})
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&pattern, "pattern", "p", pattern, "Only print names matching the pattern")

	return cmd
}

func newListFields() *cobra.Command {
	var pattern string

	cmd := &cobra.Command{
		Use:   "fields",
		Short: "List fields of the struct",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			filename, typeName := args[0], args[1]

			patternRegexp, err := compilePattern(pattern)
			if err != nil {
				return err
			}

			inspectFiles([]string{filename}, func(f *ast.File) error {
				for _, decl := range f.Decls {
					genDecl, ok := decl.(*ast.GenDecl)
					if !ok || genDecl.Tok != token.TYPE {
						continue
					}

					for _, spec := range genDecl.Specs {
						typeSpec, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}

						if typeName == typeSpec.Name.Name {
							var fields *ast.FieldList
							switch specType := typeSpec.Type.(type) {
							case *ast.StructType:
								fields = specType.Fields
							case *ast.InterfaceType:
								fields = specType.Methods
							default:
								continue
							}
							for _, fields := range fields.List {
								for _, nameIdent := range fields.Names {
									name := nameIdent.Name
									if matchPattern(patternRegexp, name) {
										fmt.Println(name)
									}
								}
							}
						}
					}
				}
				return nil
			})
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&pattern, "pattern", "p", pattern, "Only print names matching the pattern")

	return cmd
}
