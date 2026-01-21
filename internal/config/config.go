package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Handbooks struct {
		logLevel slog.Level `toml:"logLevel"`
	} `toml:"handbooks"`
	Database struct {
		Host     string `toml:"host"`
		Port     int    `toml:"port"`
		User     string `toml:"user"`
		Password string `toml:"password"`
		Name     string `toml:"name"`
		SslMode  string `toml:"sslmode"`
		URL      string
	} `toml:"database"`
	Server struct {
		Host         string `toml:"host"`
		Port         int    `toml:"port"`
		ReadTimeout  string `toml:"readTimeout"`
		WriteTimeout string `toml:"writeTimeout"`
		IdleTimeout  string `toml:"idleTimeout"`
		URL          string

		readTimeoutDur  time.Duration
		writeTimeoutDur time.Duration
		idleTimeoutDur  time.Duration
	} `toml:"server"`
}

func (c *Config) DatabaseURL() string         { return c.Database.URL }
func (c *Config) ServerURL() string           { return c.Server.URL }
func (c *Config) ReadTimeout() time.Duration  { return c.Server.readTimeoutDur }
func (c *Config) WriteTimeout() time.Duration { return c.Server.writeTimeoutDur }
func (c *Config) IdleTimeout() time.Duration  { return c.Server.idleTimeoutDur }

// NewConfig - загружает и валидирует конфигурацию
func NewConfig(context context.Context, configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		slog.ErrorContext(context, "Error read config.toml file", slog.String("error", err.Error()))
		return nil, err
	}

	var cfg Config
	if _, tomlErr := toml.Decode(string(data), &cfg); tomlErr != nil {
		slog.ErrorContext(context, "Error decode config.toml file", slog.String("error", tomlErr.Error()))
		return nil, tomlErr
	}

	if err := cfg.parseTimeouts(); err != nil {
		slog.ErrorContext(context, "Error parse timeouts", slog.String("error", err.Error()))
		return nil, err
	}

	cfg.Server.URL = fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	cfg.Database.URL = cfg.makePostgresURL()

	slog.InfoContext(context, "Config is loaded")
	return &cfg, nil
}

// parseTimeouts — парсит строки в time.Duration с валидацией
func (c *Config) parseTimeouts() error {
	var err error

	c.Server.readTimeoutDur, err = time.ParseDuration(c.Server.ReadTimeout)
	if err != nil {
		return fmt.Errorf("invalid readTimeout %q: %w", c.Server.ReadTimeout, err)
	}

	if c.Server.readTimeoutDur <= 0 {
		return fmt.Errorf("readTimeout must be positive, got %v", c.Server.readTimeoutDur)
	}

	c.Server.writeTimeoutDur, err = time.ParseDuration(c.Server.WriteTimeout)
	if err != nil {
		return fmt.Errorf("invalid writeTimeout %q: %w", c.Server.WriteTimeout, err)
	}

	if c.Server.writeTimeoutDur <= 0 {
		return fmt.Errorf("writeTimeout must be positive, got %v", c.Server.writeTimeoutDur)
	}

	c.Server.idleTimeoutDur, err = time.ParseDuration(c.Server.IdleTimeout)
	if err != nil {
		return fmt.Errorf("invalid idleTimeout %q: %w", c.Server.IdleTimeout, err)
	}

	if c.Server.idleTimeoutDur <= 0 {
		return fmt.Errorf("idleTimeout must be positive, got %v", c.Server.idleTimeoutDur)
	}

	return nil
}

// MakePostgresURL - functions for generate postgresURL in config
func (c *Config) makePostgresURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.Name, c.Database.SslMode)
}
