package main

import (
	"context"
	"handbooks/internal/config"
	"handbooks/internal/database"
	handlers "handbooks/internal/handler"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang-cz/devslog"
	_ "github.com/jackc/pgx/v5/stdlib"
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

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	NewDevLogger() // Инициализация логирования

	cfg, err := config.NewConfig(ctx, "configs/config.toml")
	if err != nil {
		slog.Error("Не удалось загрузить конфигурацию", "error", err)
		os.Exit(1)
	}

	slog.SetLogLoggerLevel(cfg.Handbooks.LogLevel)

	postgres, err := database.NewDatabase(ctx, cfg.Database.URL)
	if err != nil {
		slog.Error("Ошибка подключения к БД postgres", "err", err)
		os.Exit(1)
	}
	defer postgres.Close(ctx)

	redis, err := database.NewRedisConnection(ctx, cfg)
	if err != nil {
		slog.Error("Ошибка подключения к БД redis", "err", err)
		os.Exit(1)
	}
	defer redis.Close()

	if err := database.RunMigrations(ctx, cfg.Database.URL); err != nil {
		slog.Error("Ошибка миграций", "err", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- handlers.NewServer(postgres, redis, cfg).Run()
	}()

	select {
	case sig := <-sigChan:
		slog.Info("Получен сигнал завершения", "signal", sig.String())
	case err := <-serverErr:
		slog.Error("Сервер упал", "error", err)
		os.Exit(1)
	}

	cancel()

	slog.Info("Приложение остановлено")
}
