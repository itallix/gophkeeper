package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gophkeeper.com/internal/client/grpc"
	"gophkeeper.com/internal/client/jwt"
)

type Config struct {
	ServerURL string `mapstructure:"server_url"`
	TokenFile string `mapstructure:"token_file"`
}

var (
	cfgFile       string
	config        Config
	client        *grpc.GophkeeperClient
	tokenProvider *jwt.TokenProvider
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
	tokenProvider = jwt.NewTokenProvider(config.TokenFile)
	client, err = grpc.NewGophkeeperClient(config.ServerURL, tokenProvider)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v\n", err)
	}
}

func GetConfig() *Config {
	return &config
}

func GetTokenProvider() *jwt.TokenProvider {
	return tokenProvider
}
