package client

import (
	"github.com/spf13/cobra"

	"gophkeeper.com/internal/client/cmd"
)

func Execute() error {
	cobra.OnInitialize(cmd.InitConfig)

	rootCmd := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper secret manager client",
	}
	config := cmd.GetConfig()
	rootCmd.PersistentFlags().StringVar(&config.TokenFile, "config", "", "config file (default is $HOME/.gophkeeper.yaml)")

	rootCmd.AddCommand(
		cmd.NewUserCmd(),
		cmd.NewLoginCmd(),
		cmd.NewCardCmd(),
		cmd.NewNoteCmd(),
	)

	return rootCmd.Execute()
}
