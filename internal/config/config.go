package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port       int
	LogLevel   string
	Database   Database
	Auth       Auth
	Embeddings Embeddings
}

type Database struct {
	URL      string
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

type Auth struct {
	TokenSecret string
}

type Embeddings struct {
	APIKey     string
	Model      string
	Dimensions int
}

func Load() (Config, error) {
	port, err := loadPort()
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Port:     port,
		LogLevel: getEnv("LOG_LEVEL", "debug"),
		Database: Database{
			URL:      os.Getenv("DATABASE_URL"),
			Host:     os.Getenv("BLUEPRINT_DB_HOST"),
			Port:     os.Getenv("BLUEPRINT_DB_PORT"),
			Name:     os.Getenv("BLUEPRINT_DB_DATABASE"),
			User:     os.Getenv("BLUEPRINT_DB_USERNAME"),
			Password: os.Getenv("BLUEPRINT_DB_PASSWORD"),
		},
		Auth: Auth{
			TokenSecret: getEnv("JWT_SECRET", "local-development-secret"),
		},
		Embeddings: Embeddings{
			APIKey:     os.Getenv("OPENAI_API_KEY"),
			Model:      getEnv("EMBEDDING_MODEL", "text-embedding-3-small"),
			Dimensions: getEnvInt("EMBEDDING_DIMENSIONS", 1536),
		},
	}

	if err := cfg.Database.validate(); err != nil {
		return Config{}, err
	}
	if err := cfg.Auth.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func loadPort() (int, error) {
	value := getEnv("PORT", "8080")
	port, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid PORT %q: %w", value, err)
	}
	return port, nil
}

func (d Database) validate() error {
	if d.URL != "" {
		return nil
	}

	required := map[string]string{
		"BLUEPRINT_DB_HOST":     d.Host,
		"BLUEPRINT_DB_PORT":     d.Port,
		"BLUEPRINT_DB_DATABASE": d.Name,
		"BLUEPRINT_DB_USERNAME": d.User,
		"BLUEPRINT_DB_PASSWORD": d.Password,
	}
	for name, value := range required {
		if value == "" {
			return fmt.Errorf("missing required database config: %s", name)
		}
	}
	return nil
}

func (a Auth) validate() error {
	if a.TokenSecret == "" {
		return fmt.Errorf("missing required auth config: JWT_SECRET")
	}
	return nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
