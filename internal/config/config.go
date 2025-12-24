package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config holds scan configuration
type Config struct {
	HashAlgorithm    string
	WorkerCount      int
	ProgressInterval int // seconds
	DatabasePath     string
}

// LoadConfig loads configuration from file and environment
func LoadConfig() (*Config, error) {
	viper.SetDefault("scan.hash_algorithm", "sha512")
	viper.SetDefault("scan.concurrent_hashers", 4)
	viper.SetDefault("scan.progress_interval", "10s")
	viper.SetDefault("server.database.sqlite.name", "./dupedb.db")

	// Load from config file
	viper.SetConfigName(".dupectl")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath(".")

	// Environment variables override
	viper.SetEnvPrefix("DUPECTL")
	viper.AutomaticEnv()

	// Ignore error if config file doesn't exist
	_ = viper.ReadInConfig()

	// Parse progress_interval as duration
	progressDuration := viper.GetDuration("scan.progress_interval")
	progressSeconds := int(progressDuration / time.Second)

	return &Config{
		HashAlgorithm:    viper.GetString("scan.hash_algorithm"),
		WorkerCount:      viper.GetInt("scan.concurrent_hashers"),
		ProgressInterval: progressSeconds,
		DatabasePath:     viper.GetString("server.database.sqlite.name"),
	}, nil
}
