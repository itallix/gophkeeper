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
			resp, err := client.List(context.Background(), &pb.ListRequest{
				Type: pb.DataType_DATA_TYPE_LOGIN,
			})
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

			resp, err := client.Get(context.Background(), &pb.GetRequest{
				Type: pb.DataType_DATA_TYPE_LOGIN,
				Path: path,
			})
			if err != nil {
				fmt.Printf("Failed to retrieve login data: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Login: %s\n", resp.GetData().GetLogin().Login)
			fmt.Printf("Password: %s\n", resp.GetData().GetLogin().Password)
			fmt.Printf("Created at: %s\n", resp.GetData().GetBase().GetCreatedAt())
			fmt.Printf("Created by: %s\n", resp.GetData().GetBase().GetCreatedBy())
			fmt.Printf("Metadata: %s\n", resp.GetData().GetBase().GetMetadata())
		},
	}
	getCmd.Flags().StringP("path", "p", "", "Login path")
	_ = getCmd.MarkFlagRequired("path")

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new login secret",
		Run: func(cmd *cobra.Command, _ []string) {
			path, _ := cmd.Flags().GetString("path")

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

			resp, err := client.Create(context.Background(), &pb.CreateRequest{
				Data: &pb.TypedData{
					Type: pb.DataType_DATA_TYPE_LOGIN,
					Base: &pb.Metadata{
						Path: path,
					},
					Data: &pb.TypedData_Login{
						Login: &pb.LoginData{
							Login:    login,
							Password: string(password),
						},
					},
				},
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

			resp, err := client.Delete(context.Background(), &pb.DeleteRequest{
				Type: pb.DataType_DATA_TYPE_LOGIN,
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
