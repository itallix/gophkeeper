package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

//go:embed build_info.tpl
var buildInfoTpl string

func NewBuildCmd(version, date, commit string) *cobra.Command {
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build information",
		Run: func(cmd *cobra.Command, _ []string) {
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
				fmt.Printf("Error printing build info: %s\n", err)
				os.Exit(1)
			}
		
			err = tmpl.Execute(os.Stdout, data)
			if err != nil {
				fmt.Printf("Error printing build info: %s\n", err)
				os.Exit(1)
			}			
		},
	}

	return buildCmd
}
