package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/kv1sidisi/skrepka/internal/service/pact"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/kv1sidisi/skrepka/internal/config"
	"github.com/kv1sidisi/skrepka/internal/handler"
	"github.com/kv1sidisi/skrepka/internal/logger"
	"github.com/kv1sidisi/skrepka/internal/models"
	"github.com/kv1sidisi/skrepka/internal/service/auth"
	"github.com/kv1sidisi/skrepka/internal/storage"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Logger setup
	writer, err := logger.SetupWriter(cfg.LogPath)
	if err != nil {
		slog.Error("failed to setup log writer", "error", err)
		os.Exit(1)
	}
	log := logger.SetupLogger(cfg.Env, writer)

	log.Info("app version: 0.0.1")

	// Migrations setup
	log.Info("starting migrations")
	migrationDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	m, err := migrate.New(
		"file://migrations",
		migrationDSN,
	)
	if err != nil {
		log.Error("cannot create new migrate instance", "error", err)
		os.Exit(1)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error("failed to apply migrations", "error", err)
		os.Exit(1)
	}
	log.Info("migrations applied successfully")

	// Storage setup
	db, err := storage.New(ctx, cfg)
	if err != nil {
		log.Error("failed to init storage", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	userRepo := db.UserRepository()
	pactRepo := db.PactRepository()

	// Setup authentication providers.
	// We create provider instances here and register them.
	googleAuth := auth.NewGoogleAuthenticator(cfg.GoogleClientID)
	providerRegistry := auth.Authenticators{
		models.ProviderGoogle: googleAuth,
	}

	// Authentication service
	authService, err := auth.NewAuthService(userRepo, log, cfg.TokenTTL, cfg.JWTSecret, providerRegistry)
	if err != nil {
		log.Error("failed to create auth service", "error", err)
		os.Exit(1)
	}

	//Pact CRUD service
	pactService := pact.NewService(log, pactRepo)

	// Handlers setup
	healthHandler := handler.NewHealthHandler(log)
	oidcAuthHandler := handler.NewOIDCHandler(log, authService)
	pactHandler := handler.NewPactHandler(log, pactService)

	jwtAuth := handler.JwtAuthMiddleware(log, cfg.JWTSecret)

	// HTTP Server setup
	mux := http.NewServeMux()
	mux.Handle("/health", healthHandler)
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/api/v1/auth/oidc", oidcAuthHandler)
	mux.Handle("/api/v1/pacts/", jwtAuth(pactHandler))

	log.Info("starting server", "address", cfg.Address)

	server := &http.Server{
		Addr:         cfg.Address,
		Handler:      mux,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
