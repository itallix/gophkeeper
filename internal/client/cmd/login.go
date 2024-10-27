package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	pb "gophkeeper.com/pkg/generated/api/proto/v1"
)

func NewLoginCmd() *cobra.Command {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login management commands",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available logins",
		Run: func(cmd *cobra.Command, _ []string) {
			resp, err := client.ListLogins(context.Background(), &pb.ListLoginRequest{})
			if err != nil {
				fmt.Printf("Error listing logins: %v\n", err)
				os.Exit(1)
			}
			for _, name := range resp.GetSecrets() {
				fmt.Println(name)
			}
		},
	}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve login data by path",
		Run: func(cmd *cobra.Command, _ []string) {
			path, _ := cmd.Flags().GetString("path")

			resp, err := client.GetLogin(context.Background(), &pb.GetLoginRequest{
				Path: path,
			})
			if err != nil {
				fmt.Printf("Failed to retrieve login data: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Login: %s\n", resp.GetLogin())
			fmt.Printf("Password: %s\n", resp.GetPassword())
			fmt.Printf("Created at: %s\n", resp.GetCreatedAt())
			fmt.Printf("Created by: %s\n", resp.GetCreatedBy())
			fmt.Printf("Password: %s\n", resp.GetMetadata())
		},
	}
	getCmd.Flags().StringP("path", "p", "", "Login path")
	_ = getCmd.MarkFlagRequired("path")

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new login secret",
		Run: func(cmd *cobra.Command, _ []string) {
			path, _ := cmd.Flags().GetString("path")

			// Read password securely
			login, err := promptString("Enter login: ")
			if err != nil {
				fmt.Printf("\nFailed to read login: %v\n", err)
				os.Exit(1)
			}
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

			resp, err := client.CreateLogin(context.Background(), &pb.CreateLoginRequest{
				Path:     path,
				Login:    login,
				Password: string(password),
			})
			if err != nil {
				fmt.Printf("Failed to create a new login entry: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(resp.GetMessage())
		},
	}
	createCmd.Flags().StringP("path", "p", "", "Login path")
	_ = createCmd.MarkFlagRequired("path")

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete existing login entry",
		Run: func(cmd *cobra.Command, _ []string) {
			path, _ := cmd.Flags().GetString("path")

			resp, err := client.DeleteLogin(context.Background(), &pb.DeleteLoginRequest{
				Path: path,
			})
			if err != nil {
				fmt.Printf("Error deleting login: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(resp.GetMessage())
		},
	}
	deleteCmd.Flags().StringP("path", "p", "", "Login path")
	_ = deleteCmd.MarkFlagRequired("path")

	loginCmd.AddCommand(listCmd, getCmd, createCmd, deleteCmd)

	return loginCmd
}
