package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	pb "github.com/itallix/gophkeeper/pkg/generated/api/proto/v1"
)

func NewLoginCmd() *cobra.Command {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login management commands",
	}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve login data by path",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, _ := cmd.Flags().GetString("path")

			resp, err := client.Get(context.Background(), &pb.GetRequest{
				Type: pb.DataType_DATA_TYPE_LOGIN,
				Path: path,
			})
			if err != nil {
				return fmt.Errorf("failed to retrieve login data: %w", err)
			}
			cmd.Printf("Login: %s\n", resp.GetData().GetLogin().GetLogin())
			cmd.Printf("Password: %s\n", resp.GetData().GetLogin().GetPassword())
			cmd.Printf("Created at: %s\n", resp.GetData().GetBase().GetCreatedAt())
			cmd.Printf("Created by: %s\n", resp.GetData().GetBase().GetCreatedBy())
			cmd.Printf("Metadata: %s\n", resp.GetData().GetBase().GetMetadata())
			return nil
		},
	}
	getCmd.Flags().StringP("path", "p", "", "Login path")
	_ = getCmd.MarkFlagRequired("path")

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new login secret",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, _ := cmd.Flags().GetString("path")
			reader := bufio.NewReader(cmd.InOrStdin())

			login, err := promptString(cmd, reader, "Enter login: ")
			if err != nil {
				return fmt.Errorf("failed to read login: %w", err)
			}

			password, err := promptPassword(cmd, reader, "Enter password: ")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			confirm, err := promptPassword(cmd, reader, "Confirm password: ")
			if err != nil {
				return fmt.Errorf("failed to read password confirmation: %w", err)
			}
			cmd.Println()

			if password != confirm {
				return errors.New("passwords don't match")
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
							Password: password,
						},
					},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to create a new login entry: %w", err)
			}
			cmd.Println(resp.GetMessage())
			return nil
		},
	}
	createCmd.Flags().StringP("path", "p", "", "Login path")
	_ = createCmd.MarkFlagRequired("path")

	listCmd := NewListCmd("login", "List available logins", pb.DataType_DATA_TYPE_LOGIN)
	deleteCmd := NewDeleteCmd("login", "Delete existing login entry", pb.DataType_DATA_TYPE_LOGIN)

	loginCmd.AddCommand(listCmd, getCmd, createCmd, deleteCmd)

	return loginCmd
}
