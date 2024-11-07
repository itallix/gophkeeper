package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gophkeeper.com/internal/client/jwt"
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			login, _ := cmd.Flags().GetString("login")
			reader := bufio.NewReader(cmd.InOrStdin())

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

			resp, err := client.Register(context.Background(), &pb.RegisterRequest{
				Login:    login,
				Password: password,
			})
			if err != nil {
				return fmt.Errorf("dailed to register: %w", err)
			}
			tokenData := jwt.NewToken(resp.GetAccessToken(), resp.GetRefreshToken())
			if err = tokenProvider.SaveToken(tokenData); err != nil {
				return fmt.Errorf("failed to save token: %w", err)
			}
			cmd.Printf("User with login=%s successfully registered", login)
			return nil
		},
	}
	registerCmd.Flags().StringP("login", "l", "", "Username")
	_ = registerCmd.MarkFlagRequired("login")

	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Login to the service",
		RunE: func(cmd *cobra.Command, _ []string) error {
			login, _ := cmd.Flags().GetString("login")
			reader := bufio.NewReader(cmd.InOrStdin())

			password, err := promptPassword(cmd, reader, "Enter password: ")
			cmd.Println()
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}

			resp, err := client.Login(context.Background(), &pb.LoginRequest{
				Login:    login,
				Password: password,
			})
			if err != nil {
				return fmt.Errorf("failed to login: %w", err)
			}

			tokenData := jwt.NewToken(resp.GetAccessToken(), resp.GetRefreshToken())
			if err = tokenProvider.SaveToken(tokenData); err != nil {
				return fmt.Errorf("failed to save token: %w", err)
			}
			cmd.Printf("Successfully logged in as %s", login)
			return nil
		},
	}
	authCmd.Flags().StringP("login", "l", "", "Username")
	_ = authCmd.MarkFlagRequired("login")

	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from the service",
		RunE: func(cmd *cobra.Command, _ []string) error {
			err := os.Remove(config.TokenFile)
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove token: %w", err)
			}
			cmd.Print("Successfully logged out")
			return nil
		},
	}
	userCmd.AddCommand(registerCmd, authCmd, logoutCmd)

	return userCmd
}
