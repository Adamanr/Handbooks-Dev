package main

import (
	"context"
	"handbooks/internal/config"
	"handbooks/internal/database"
	"handbooks/internal/handlers"
	"log/slog"
	"os"

	"github.com/golang-cz/devslog"
)

func NewDevLogger() {
	opts := &devslog.Options{
		MaxSlicePrintSize: 4,
		SortKeys:          true,
		TimeFormat:        "15:04:05.000",
		NewLineAfterLog:   true,
		DebugColor:        devslog.Magenta,
		StringerFormatter: true,
	}

	handler := devslog.NewHandler(os.Stdout, opts)
	logger := slog.New(handler)

	slog.SetDefault(logger)
}

// init - Инициализация приложения
func init() {
	NewDevLogger()
}

func main() {
	ctx := context.Background()

	cfg, err := config.NewConfig(ctx, "configs/config.toml")
	if err != nil {
		slog.ErrorContext(ctx, "Error read config.toml file", slog.String("error", err.Error()))
		return
	}

	database, err := database.NewDatabase(ctx, cfg.DatabaseURL())
	if err != nil {
		slog.ErrorContext(ctx, "Error create new database connections!", slog.String("Error", err.Error()))
		return
	}

	handlers.NewServer(database, cfg).Run()
}
