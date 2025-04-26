package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv  string
	AppPort string

	DBHost               string
	DBPort               int
	DBUser               string
	DBPassword           string
	DBName               string
	DBMaxOpenConnections int
	DBMaxIdleConnections int

	JWT struct {
		SecretKey     string
		TokenDuration int
	}

	SMTP struct {
		Host      string
		Port      int
		Username  string
		Password  string
		FromEmail string
	}

	CloudflareR2BucketName string
	CloudflareR2APIKey     string
	CloudflareR2APISecret  string
	CloudflareR2Token      string
	CloudflareR2AccountID  string
	CloudflareR2PublicURL  string

	Server struct {
		Port int
	}

	FrontendURL string
}

func findRootDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".env")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}

func setDefaults() {
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
	viper.SetDefault("FRONTEND_URL", "http://localhost:3000")
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("No config file found at %s, using defaults", path)
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config

	config.AppEnv = viper.GetString("APP_ENV")
	config.AppPort = viper.GetString("APP_PORT")

	config.DBHost = viper.GetString("DATABASE_HOST")
	config.DBPort = viper.GetInt("DATABASE_PORT")
	config.DBUser = viper.GetString("DATABASE_USER")
	config.DBPassword = viper.GetString("DATABASE_PASSWORD")
	config.DBName = viper.GetString("DATABASE_NAME")
	config.DBMaxOpenConnections = viper.GetInt("DATABASE_MAX_OPEN_CONNECTIONS")
	config.DBMaxIdleConnections = viper.GetInt("DATABASE_MAX_IDLE_CONNECTIONS")

	config.JWT.SecretKey = viper.GetString("JWT_SECRET_KEY")
	config.JWT.TokenDuration = viper.GetInt("JWT_TOKEN_DURATION")

	config.SMTP.Host = viper.GetString("SMTP_HOST")
	config.SMTP.Port = viper.GetInt("SMTP_PORT")
	config.SMTP.Username = viper.GetString("SMTP_USERNAME")
	config.SMTP.Password = viper.GetString("SMTP_PASSWORD")
	config.SMTP.FromEmail = viper.GetString("SMTP_FROM_EMAIL")

	config.CloudflareR2BucketName = viper.GetString("CLOUDFLARE_R2_BUCKET_NAME")
	config.CloudflareR2APIKey = viper.GetString("CLOUDFLARE_R2_API_KEY")
	config.CloudflareR2APISecret = viper.GetString("CLOUDFLARE_R2_API_SECRET")
	config.CloudflareR2Token = viper.GetString("CLOUDFLARE_R2_TOKEN")
	config.CloudflareR2AccountID = viper.GetString("CLOUDFLARE_R2_ACCOUNT_ID")
	config.CloudflareR2PublicURL = viper.GetString("CLOUDFLARE_R2_PUBLIC_URL")

	config.FrontendURL = viper.GetString("FRONTEND_URL")

	port, err := strconv.Atoi(strings.TrimSpace(config.AppPort))
	if err != nil {
		return nil, fmt.Errorf("invalid APP_PORT: %v", err)
	}
	config.Server.Port = port

	return &config, nil
}

func NewConfig() *Config {
	rootDir := findRootDir()
	configPath := filepath.Join(rootDir, ".env")

	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	return config
}

func (c *Config) GetAppPort() string {
	return c.AppPort
}

func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}
