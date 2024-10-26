package client

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	ServerURL string `mapstructure:"server_url"`
	TokenFile string `mapstructure:"token_file"`
}

var (
	cfgFile string
	config  Config
)

func initConfig() {
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
}
