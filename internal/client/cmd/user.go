package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	pb "gophkeeper.com/pkg/generated/api/proto/v1"
)

func NewUserCmd() *cobra.Command {
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "User management commands",
	}

	registerCmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		Run: func(cmd *cobra.Command, _ []string) {
			login, _ := cmd.Flags().GetString("login")

			// Read password securely
			fmt.Print("Enter password: ")
			password, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				fmt.Printf("\nFailed to read password: %v\n", err)
				os.Exit(1)
			}
			fmt.Print("\nConfirm password: ")
			confirm, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				fmt.Printf("\nFailed to read password confirmation: %v\n", err)
				os.Exit(1)
			}
			fmt.Println()

			if string(password) != string(confirm) {
				fmt.Println("Passwords don't match")
				os.Exit(1)
			}

			resp, err := client.Register(context.Background(), &pb.RegisterRequest{
				Login:    login,
				Password: string(password),
			})
			if err != nil {
				fmt.Printf("Failed to register: %v\n", err)
				os.Exit(1)
			}
			err = SaveToken(resp.GetToken())
			if err != nil {
				fmt.Printf("Failed to save token: %v\n", err)
				os.Exit(1)
			}
			fmt.Print("Successfully registered")
		},
	}
	registerCmd.Flags().StringP("login", "l", "", "Username")
	_ = registerCmd.MarkFlagRequired("login")

	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Login to the service",
		Run: func(cmd *cobra.Command, _ []string) {
			login, _ := cmd.Flags().GetString("login")

			fmt.Print("Enter password: ")
			password, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				fmt.Printf("Failed to read password: %v\n", err)
				os.Exit(1)
			}

			resp, err := client.Login(context.Background(), &pb.LoginRequest{
				Login:    login,
				Password: string(password),
			})
			if err != nil {
				fmt.Printf("Failed to login: %v\n", err)
				os.Exit(1)
			}

			err = SaveToken(resp.GetToken())
			if err != nil {
				fmt.Printf("Failed to save token: %v\n", err)
				os.Exit(1)
			}
			fmt.Print("Successfully logged in")
		},
	}
	authCmd.Flags().StringP("login", "l", "", "Username")
	_ = authCmd.MarkFlagRequired("login")

	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from the service",
		Run: func(_ *cobra.Command, _ []string) {
			err := os.Remove(config.TokenFile)
			if err != nil && !os.IsNotExist(err) {
				fmt.Printf("Failed to remove token: %v\n", err)
				os.Exit(1)
			}
			fmt.Print("Successfully logged out")
		},
	}
	userCmd.AddCommand(registerCmd, authCmd, logoutCmd)

	return userCmd
}
