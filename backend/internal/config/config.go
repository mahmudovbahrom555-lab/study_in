// Package config загружает и валидирует конфигурацию приложения.
// Источники: .env файл и переменные окружения.
// Переменные окружения имеют приоритет над .env.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config содержит всю конфигурацию приложения.
// Каждая секция — отдельная подсистема.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	S3       S3Config
	SMS      SMSConfig
	FCM      FCMConfig
	Sentry   SentryConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port    string
	Env     string // development, staging, production
	Version string
}

func (s ServerConfig) IsDevelopment() bool {
	return s.Env == "development"
}

func (s ServerConfig) IsProduction() bool {
	return s.Env == "production"
}

type DatabaseConfig struct {
	URL      string
	MaxConns int
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	Region    string
}

type SMSConfig struct {
	Provider      string // mock, eskiz
	EskizEmail    string
	EskizPassword string
	EskizFrom     string
}

type FCMConfig struct {
	CredentialsPath string
	ProjectID       string
}

type SentryConfig struct {
	DSN string
}

type LogConfig struct {
	Level  string // debug, info, warn, error
	Format string // text, json
}

// Load читает .env (если есть) и переменные окружения, возвращает Config.
func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AddConfigPath("./backend")
	v.AddConfigPath("../")

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Значения по умолчанию
	v.SetDefault("SERVER_PORT", "8080")
	v.SetDefault("SERVER_ENV", "development")
	v.SetDefault("SERVER_VERSION", "0.1.0")
	v.SetDefault("DATABASE_MAX_CONNS", 25)
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("JWT_ACCESS_TTL_MINUTES", 15)
	v.SetDefault("JWT_REFRESH_TTL_DAYS", 30)
	v.SetDefault("S3_USE_SSL", false)
	v.SetDefault("S3_REGION", "us-east-1")
	v.SetDefault("SMS_PROVIDER", "mock")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "text")

	// .env не обязателен — переменные окружения тоже работают
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:    v.GetString("SERVER_PORT"),
			Env:     v.GetString("SERVER_ENV"),
			Version: v.GetString("SERVER_VERSION"),
		},
		Database: DatabaseConfig{
			URL:      v.GetString("DATABASE_URL"),
			MaxConns: v.GetInt("DATABASE_MAX_CONNS"),
		},
		Redis: RedisConfig{
			Addr:     v.GetString("REDIS_ADDR"),
			Password: v.GetString("REDIS_PASSWORD"),
			DB:       v.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			AccessSecret:  v.GetString("JWT_ACCESS_SECRET"),
			RefreshSecret: v.GetString("JWT_REFRESH_SECRET"),
			AccessTTL:     time.Duration(v.GetInt("JWT_ACCESS_TTL_MINUTES")) * time.Minute,
			RefreshTTL:    time.Duration(v.GetInt("JWT_REFRESH_TTL_DAYS")) * 24 * time.Hour,
		},
		S3: S3Config{
			Endpoint:  v.GetString("S3_ENDPOINT"),
			AccessKey: v.GetString("S3_ACCESS_KEY"),
			SecretKey: v.GetString("S3_SECRET_KEY"),
			Bucket:    v.GetString("S3_BUCKET"),
			UseSSL:    v.GetBool("S3_USE_SSL"),
			Region:    v.GetString("S3_REGION"),
		},
		SMS: SMSConfig{
			Provider:      v.GetString("SMS_PROVIDER"),
			EskizEmail:    v.GetString("SMS_ESKIZ_EMAIL"),
			EskizPassword: v.GetString("SMS_ESKIZ_PASSWORD"),
			EskizFrom:     v.GetString("SMS_ESKIZ_FROM"),
		},
		FCM: FCMConfig{
			CredentialsPath: v.GetString("FCM_CREDENTIALS_PATH"),
			ProjectID:       v.GetString("FCM_PROJECT_ID"),
		},
		Sentry: SentryConfig{
			DSN: v.GetString("SENTRY_DSN"),
		},
		Log: LogConfig{
			Level:  v.GetString("LOG_LEVEL"),
			Format: v.GetString("LOG_FORMAT"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWT.AccessSecret == "" || len(c.JWT.AccessSecret) < 32 {
		return fmt.Errorf("JWT_ACCESS_SECRET must be at least 32 chars")
	}
	if c.JWT.RefreshSecret == "" || len(c.JWT.RefreshSecret) < 32 {
		return fmt.Errorf("JWT_REFRESH_SECRET must be at least 32 chars")
	}
	return nil
}
