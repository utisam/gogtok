package command

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"

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

func newListFuncs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "funcs",
		Short: "List functions of the file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			inspectFiles(args, func(f *ast.File) error {
				for _, decl := range f.Decls {
					fnDecl, ok := decl.(*ast.FuncDecl)
					if !ok {
						continue
					}
					fmt.Println(fnDecl.Name.Name)
				}
				return nil
			})
			return nil
		},
	}
	return cmd
}

func newListValues() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "values",
		Short: "List variables of the file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
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

						for _, name := range valSpec.Names {
							fmt.Println(name)
						}
					}
				}
				return nil
			})
			return nil
		},
	}

	return cmd
}

func newListTypes() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "types",
		Short: "List types of the file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
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

						fmt.Println(typeSpec.Name.Name)
					}
				}
				return nil
			})
			return nil
		},
	}

	return cmd
}

func newListFields() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fields",
		Short: "List fields of the struct",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			filename, typeName := args[0], args[1]

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
								for _, name := range fields.Names {
									fmt.Println(name.Name)
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

	return cmd
}
