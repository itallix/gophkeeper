package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gophkeeper.com/internal/client/grpc"
)

type Config struct {
	ServerURL string `mapstructure:"server_url"`
	TokenFile string `mapstructure:"token_file"`
}

var (
	cfgFile string
	config  Config
	client  *grpc.GophkeeperClient
)

func InitConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".gophkeeper"
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".gophkeeper")
	}

	// Set defaults
	viper.SetDefault("server_url", "localhost:8081")
	viper.SetDefault("token_file", filepath.Join(os.TempDir(), ".gophkeeper_token"))

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Failed to parse config: %v\n", err)
	}

	var err error
	token, _ := LoadToken()
	client, err = grpc.NewGophkeeperClient(config.ServerURL, token)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v\n", err)
	}
}

// Token storage with file permissions.
func SaveToken(token string) error {
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
	err = os.WriteFile(config.TokenFile, jsonData, 0600)
	if err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

func LoadToken() (string, error) {
	data, err := os.ReadFile(config.TokenFile)
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

func GetConfig() *Config {
	return &config
}
