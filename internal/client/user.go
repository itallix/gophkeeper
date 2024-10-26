package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	pb "gophkeeper.com/pkg/generated/api/proto/v1"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "User management commands",
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user",
	Run: func(cmd *cobra.Command, _ []string) {
		login, _ := cmd.Flags().GetString("login")

		// Read password securely
		log.Print("Enter password: ")
		password, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("\nFailed to read password: %v\n", err)
		}
		log.Print("\nConfirm password: ")
		confirm, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("\nFailed to read password confirmation: %v\n", err)
		}
		log.Println()

		if string(password) != string(confirm) {
			log.Fatal("Passwords don't match")
		}

		client, err := NewGophkeeperClient(config.ServerURL)
		if err != nil {
			log.Fatalf("Failed to connect to the server: %v\n", err)
		}
		resp, err := client.Register(context.Background(), &pb.RegisterRequest{
			Login:    login,
			Password: string(password),
		})
		if err != nil {
			log.Fatalf("Failed to register: %v\n", err)
		}
		err = SaveToken(config.TokenFile, resp.GetToken())
		if err != nil {
			log.Fatalf("Failed to save token: %v\n", err)
		}
		log.Print("Successfully registered")
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to the service",
	Run: func(cmd *cobra.Command, _ []string) {
		login, _ := cmd.Flags().GetString("login")

		log.Print("Enter password: ")
		password, err := term.ReadPassword(int(os.Stdin.Fd()))
		log.Println()
		if err != nil {
			log.Fatalf("Failed to read password: %v\n", err)
		}

		client, err := NewGophkeeperClient(config.ServerURL)
		if err != nil {
			log.Fatalf("Failed to connect to the server: %v\n", err)
		}
		resp, err := client.Login(context.Background(), &pb.LoginRequest{
			Login:    login,
			Password: string(password),
		})
		if err != nil {
			log.Fatalf("Failed to login: %v\n", err)
		}

		err = SaveToken(config.TokenFile, resp.GetToken())
		if err != nil {
			log.Fatalf("Failed to save token: %v\n", err)
		}
		log.Print("Successfully logged in")
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from the service",
	Run: func(_ *cobra.Command, _ []string) {
		err := os.Remove(config.TokenFile)
		if err != nil && !os.IsNotExist(err) {
			log.Fatalf("Failed to remove token: %v\n", err)
		}
		log.Print("Successfully logged out")
	},
}

// Token storage with file permissions.
func SaveToken(filepath string, token string) error {
	data := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Create file with user-only read/write permissions
	err = os.WriteFile(filepath, jsonData, 0600)
	if err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

func LoadToken(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to read token file: %w", err)
	}

	var tokenData struct {
		Token string `json:"token"`
	}

	err = json.Unmarshal(data, &tokenData)
	if err != nil {
		return "", fmt.Errorf("failed to parse token file: %w", err)
	}

	return tokenData.Token, nil
}
