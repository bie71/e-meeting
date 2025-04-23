package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv  string `mapstructure:"APP_ENV"`
	AppPort string `mapstructure:"APP_PORT"`

	// Database Configuration
	DBHost               string `mapstructure:"DATABASE_HOST"`
	DBPort               int    `mapstructure:"DATABASE_PORT"`
	DBUser               string `mapstructure:"DATABASE_USER"`
	DBPassword           string `mapstructure:"DATABASE_PASSWORD"`
	DBName               string `mapstructure:"DATABASE_NAME"`
	DBMaxOpenConnections int    `mapstructure:"DATABASE_MAX_OPEN_CONNECTIONS"`
	DBMaxIdleConnections int    `mapstructure:"DATABASE_MAX_IDLE_CONNECTIONS"`

	// JWT Configuration
	JWT struct {
		SecretKey     string `mapstructure:"JWT_SECRET_KEY"`
		TokenDuration int    `mapstructure:"JWT_TOKEN_DURATION"`
	}

	// SMTP Configuration
	SMTP struct {
		Host      string `mapstructure:"SMTP_HOST"`
		Port      int    `mapstructure:"SMTP_PORT"`
		Username  string `mapstructure:"SMTP_USERNAME"`
		Password  string `mapstructure:"SMTP_PASSWORD"`
		FromEmail string `mapstructure:"SMTP_FROM_EMAIL"`
	}

	// Cloudflare R2 Configuration
	CloudflareR2BucketName string `mapstructure:"CLOUDFLARE_R2_BUCKET_NAME"`
	CloudflareR2APIKey     string `mapstructure:"CLOUDFLARE_R2_API_KEY"`
	CloudflareR2APISecret  string `mapstructure:"CLOUDFLARE_R2_API_SECRET"`
	CloudflareR2Token      string `mapstructure:"CLOUDFLARE_R2_TOKEN"`
	CloudflareR2AccountID  string `mapstructure:"CLOUDFLARE_R2_ACCOUNT_ID"`
	CloudflareR2PublicURL  string `mapstructure:"CLOUDFLARE_R2_PUBLIC_URL"`

	Server struct {
		Port int
	}
}

func findRootDir() string {
	// Start from the current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Walk up the directory tree until we find .env or reach the root
	for {
		// Check if .env exists in current directory
		if _, err := os.Stat(filepath.Join(dir, ".env")); err == nil {
			return dir
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// We've reached the root
			return "."
		}
		dir = parent
	}
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("DATABASE_HOST", "localhost")
	viper.SetDefault("DATABASE_PORT", 5432)
	viper.SetDefault("DATABASE_USER", "postgres")
	viper.SetDefault("DATABASE_PASSWORD", "postgres")
	viper.SetDefault("DATABASE_NAME", "e_metting")
	viper.SetDefault("DATABASE_MAX_OPEN_CONNECTIONS", 25)
	viper.SetDefault("DATABASE_MAX_IDLE_CONNECTIONS", 5)
	viper.SetDefault("JWT_SECRET_KEY", "your-secret-key")
	viper.SetDefault("JWT_TOKEN_DURATION", 24)
	viper.SetDefault("SMTP_HOST", "smtp.gmail.com")
	viper.SetDefault("SMTP_PORT", 587)
	viper.SetDefault("SMTP_USERNAME", "your-email@gmail.com")
	viper.SetDefault("SMTP_PASSWORD", "your-password")
	viper.SetDefault("SMTP_FROM_EMAIL", "your-email@gmail.com")
	viper.SetDefault("CLOUDFLARE_R2_BUCKET_NAME", "")
	viper.SetDefault("CLOUDFLARE_R2_API_KEY", "")
	viper.SetDefault("CLOUDFLARE_R2_API_SECRET", "")
	viper.SetDefault("CLOUDFLARE_R2_TOKEN", "")
	viper.SetDefault("CLOUDFLARE_R2_ACCOUNT_ID", "")
	viper.SetDefault("CLOUDFLARE_R2_PUBLIC_URL", "")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if we want to use defaults
			log.Printf("No config file found at %s, using defaults", path)
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Load server config
	port, err := strconv.Atoi(viper.GetString("PORT"))
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}
	config.Server.Port = port

	return &config, nil
}

func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func (c *Config) GetAppPort() string {
	return c.AppPort
}

func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

func InitConfig() *Config {
	rootDir := findRootDir()
	configPath := filepath.Join(rootDir, ".env")

	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	return config
}

func NewConfig() *Config {
	return InitConfig()
}
