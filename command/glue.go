package command

import (
	"io"
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

type GlueTemplateData struct {
	From      string
	To        string
	Converter string
	Field     string
}

const (
	assignConvertLineTemplate = `{{.To}}.{{.Field}} = {{.Converter}}({{.From}}.{{.Field}})`
	assignLineTemplate        = `{{.To}}.{{.Field}} = {{.From}}.{{.Field}}`
	convertLineTemplate       = `{{.Field}}: {{.Converter}}({{.From}}.{{.Field}})`
)

func newGlue() *cobra.Command {
	lineTemplate := `{{.Field}}: {{.From}}.{{.Field}},`
	to := ""
	converter := ""

	cmd := &cobra.Command{
		Use:   "glue",
		Short: "Generate glue code",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			from := args[0]
			fields := args[1:]

			if !cmd.Flag("line-template").Changed {
				if to != "" && converter != "" {
					lineTemplate = assignConvertLineTemplate
				} else if to != "" && converter == "" {
					lineTemplate = assignLineTemplate
				} else if to == "" && converter != "" {
					lineTemplate = convertLineTemplate
				}
			}

			tmpl, err := template.New("").Parse(lineTemplate)
			if err != nil {
				return err
			}

			for _, field := range fields {
				tmpl.Execute(os.Stdout, &GlueTemplateData{
					From:      from,
					To:        to,
					Converter: converter,
					Field:     field,
				})
				io.WriteString(os.Stdout, "\n")
			}
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&lineTemplate, "line-template", lineTemplate, "Template to render each line")
	flags.StringVar(&to, "to", to, `Shorthand for --line-template="`+assignLineTemplate+`"`)
	flags.StringVar(&converter, "converter", converter, `Shorthand for --line-template="`+convertLineTemplate+`"`)

	return cmd
}
