package client

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gophkeeper",
	Short: "GophKeeper password manager client",
}

func Execute() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gophkeeper.yaml)")

	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(registerCmd, loginCmd, logoutCmd)

	// Add flags
	registerCmd.Flags().StringP("register", "l", "", "User register")
	_ = registerCmd.MarkFlagRequired("register")

	loginCmd.Flags().StringP("login", "l", "", "User login")
	_ = loginCmd.MarkFlagRequired("login")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
