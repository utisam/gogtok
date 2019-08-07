package command

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"regexp"
	"strings"

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

func inspectFiles(filenames []string, fn func(fset *token.FileSet, f *ast.File) error) error {
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

		if err := fn(fset, f); err != nil {
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

			inspectFiles(args, func(fset *token.FileSet, f *ast.File) error {
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

			inspectFiles(args, func(fset *token.FileSet, f *ast.File) error {
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

			inspectFiles(args, func(fset *token.FileSet, f *ast.File) error {
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

type fieldInfo struct {
	name      string
	fieldType string
	tag       string
}

func newFieldInfo(name string, typeExpr ast.Expr, tagLit *ast.BasicLit) *fieldInfo {
	tag := ""
	if tagLit != nil {
		tag = strings.Trim(tagLit.Value, "`")
	}

	return &fieldInfo{
		name:      name,
		fieldType: typeString(typeExpr),
		tag:       tag,
	}
}

type fieldInfoColumn func(*fieldInfo) string

func parseColumns(ss []string) ([]fieldInfoColumn, error) {
	cols := make([]fieldInfoColumn, len(ss))
	for i, s := range ss {
		col, err := parseColumn(s)
		if err != nil {
			return nil, err
		}
		cols[i] = col
	}
	return cols, nil
}

var tagColumnPattern = regexp.MustCompile(`tag\[([^\]]+)\]`)

func parseColumn(s string) (fieldInfoColumn, error) {
	switch s {
	case "name":
		return func(fi *fieldInfo) string {
			return fi.name
		}, nil
	case "type":
		return func(fi *fieldInfo) string {
			return fi.fieldType
		}, nil
	case "tags":
		return func(fi *fieldInfo) string {
			return fi.tag
		}, nil
	}

	submatch := tagColumnPattern.FindStringSubmatch(s)
	if submatch != nil {
		return func(fi *fieldInfo) string {
			return reflect.StructTag(fi.tag).Get(submatch[1])
		}, nil
	}

	return nil, fmt.Errorf("Unknown column: %s", s)
}

func (fi *fieldInfo) toValues(cols []fieldInfoColumn) []string {
	res := make([]string, len(cols))
	for i, col := range cols {
		res[i] = col(fi)
	}
	return res
}

func defaultFieldPrinter(a ...string) {
	fmt.Println(strings.Join(a, " "))
}

func nullCharFieldPrinter(a ...string) {
	fmt.Print(strings.Join(a, "\x00"))
	fmt.Print("\x00")
}

func newListFields() *cobra.Command {
	pattern := ""
	print0 := false
	columns := []string{"name"}

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

			p := defaultFieldPrinter
			if print0 {
				p = nullCharFieldPrinter
			}

			cols, err := parseColumns(columns)
			if err != nil {
				return err
			}

			inspectFiles([]string{filename}, func(fset *token.FileSet, f *ast.File) error {
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
							for _, field := range fields.List {
								for _, nameIdent := range field.Names {
									name := nameIdent.Name
									if matchPattern(patternRegexp, name) {
										info := newFieldInfo(name, field.Type, field.Tag)
										p(info.toValues(cols)...)
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
	flags.BoolVarP(&print0, "print0", "0", print0, "Print info followed by a null character")
	flags.StringSliceVar(&columns, "columns", columns, "Columns to be output (name, type, tags, tag[key])")

	return cmd
}

func typeString(expr ast.Expr) string {
	b := &strings.Builder{}
	b.Grow(int(expr.End()) - int(expr.Pos()))
	appendExpr(b, expr)
	return b.String()
}

func appendExpr(b *strings.Builder, expr ast.Expr) {
	switch t := expr.(type) {
	case nil:
	case *ast.Ident:
		b.WriteString(t.Name)
	case *ast.SelectorExpr:
		appendExpr(b, t.X)
		b.WriteRune('.')
		appendExpr(b, t.Sel)
	case *ast.StarExpr:
		b.WriteRune('*')
		appendExpr(b, t.X)
	case *ast.ArrayType:
		b.WriteRune('[')
		appendExpr(b, t.Len)
		b.WriteRune(']')
		appendExpr(b, t.Elt)
	case *ast.StructType:
		// TODO: Fields
		b.WriteString("struct{}")
	case *ast.FuncType:
		numParams := t.Params.NumFields()
		if numParams == 0 {
			b.WriteString("func()")
		} else {
			b.WriteString("func(")
			appendFieldList(b, t.Params)
			b.WriteRune(')')
		}
		numResults := t.Results.NumFields()
		if numResults == 1 {
			b.WriteRune(' ')
			appendFieldList(b, t.Results)
		} else if numResults >= 2 {
			b.WriteString(" (")
			appendFieldList(b, t.Results)
			b.WriteRune(')')
		}
	case *ast.InterfaceType:
		// TODO: Fields
		b.WriteString("interface{}")
	case *ast.MapType:
		b.WriteString("map[")
		appendExpr(b, t.Key)
		b.WriteString("]")
		appendExpr(b, t.Value)
	case *ast.ChanType:
		if t.Dir == ast.RECV {
			b.WriteString("<-")
		}
		b.WriteString("chan")
		if t.Dir == ast.SEND {
			b.WriteString("<-")
		}
		b.WriteRune(' ')
		appendExpr(b, t.Value)
	case *ast.BasicLit:
		b.WriteString(t.Value)
	default:
		panic(fmt.Sprintf("unsupported type: %#v", expr))
	}
}

func appendFieldList(b *strings.Builder, expr *ast.FieldList) {
	for i, field := range expr.List {
		if i != 0 {
			b.WriteString(", ")
		}
		if len(field.Names) > 0 {
			for j, name := range field.Names {
				if j != 0 {
					b.WriteString(", ")
				}
				b.WriteString(name.Name)
			}
			b.WriteRune(' ')
		}
		appendExpr(b, field.Type)
	}
}
