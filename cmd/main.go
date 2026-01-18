package main

import (
	"context"
	"handbooks/internal/api"
	"handbooks/internal/config"
	"handbooks/internal/database"
	"handbooks/internal/handlers"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	config, err := config.NewConfig(logger)
	if err != nil {
		logger.Error("Error read config.toml file", slog.String("error", err.Error()))
		return
	}

	postgresURL := config.MakePostgresURL()
	database := database.NewDatabase(context.Background(), postgresURL)

	server := handlers.NewServer(database, config, logger)

	r := chi.NewMux()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	h := api.HandlerFromMux(server, r)

	s := &http.Server{
		Handler: h,
		Addr:    config.Handbooks.Server,
	}

	slog.Info("Server started!", "url", "http://localhost:3000")
	log.Fatal(s.ListenAndServe())
}
