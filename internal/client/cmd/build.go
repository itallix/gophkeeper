package cmd

import (
	_ "embed"
	"fmt"
	"text/template"

	"github.com/spf13/cobra"
)

//go:embed build_info.tpl
var buildInfoTpl string

func NewBuildCmd(version, date, commit string) *cobra.Command {
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			data := struct {
				Version string
				Date    string
				Commit  string
			}{
				Version: version,
				Date:    date,
				Commit:  commit,
			}

			tmpl, err := template.New("buildInfo").Parse(buildInfoTpl)
			if err != nil {
				return fmt.Errorf("error printing build info: %w", err)
			}

			if err = tmpl.Execute(cmd.OutOrStdout(), data); err != nil {
				return fmt.Errorf("error printing build info: %w", err)
			}
			return nil
		},
	}

	return buildCmd
}
