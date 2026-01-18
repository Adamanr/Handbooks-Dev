package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Handbooks struct {
		Server string `toml:"server"`
	} `toml:"handbooks"`
	Database struct {
		Host     string `toml:"host"`
		Port     int    `toml:"port"`
		User     string `toml:"user"`
		Password string `toml:"password"`
		Name     string `toml:"name"`
		SslMode  string `toml:"sslmode"`
	} `toml:"database"`
}

func NewConfig(logger *slog.Logger) (*Config, error) {
	data, err := os.ReadFile("configs/config.toml")
	if err != nil {
		logger.Error("Error read config.toml file", slog.String("error", err.Error()))
		return nil, err
	}

	var cfg *Config

	if _, tomlErr := toml.Decode(string(data), &cfg); tomlErr != nil {
		logger.Error("Error decode config.toml file", slog.String("error", tomlErr.Error()))
		return nil, tomlErr
	}

	logger.Info("Config is loaded")
	return cfg, nil
}

func (c *Config) MakePostgresURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.Name, c.Database.SslMode)
}
