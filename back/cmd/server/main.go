package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/build-assistant/back/config"
	dbmigrate "github.com/build-assistant/back/db"
	db "github.com/build-assistant/back/db/gen"
	"github.com/build-assistant/back/internal/ai"
	"github.com/build-assistant/back/internal/api"
	"github.com/build-assistant/back/internal/api/handlers"
	"github.com/build-assistant/back/internal/auth"
	"github.com/build-assistant/back/internal/git"
	"github.com/build-assistant/back/internal/notifications"
	"github.com/build-assistant/back/internal/services"
	"github.com/build-assistant/back/internal/workers"
	"github.com/build-assistant/back/pkg/s3"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	deps, err := setupDependencies(cfg)
	if err != nil {
		slog.Error("Failed to setup dependencies", "error", err)
		os.Exit(1)
	}

	h := handlers.NewHandlers(deps)
	router := api.NewRouter(h)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("Starting server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exited")
}

func setupDependencies(cfg *config.Config) (*handlers.Dependencies, error) {
	ctx := context.Background()

	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := dbmigrate.RunMigrations(dbPool); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	queries := db.New(dbPool)

	s3Service := s3.NewS3Service(s3.S3Config{
		S3Endpoint:     cfg.S3Endpoint,
		S3Region:       cfg.S3Region,
		S3PathStyle:    cfg.S3PathStyle,
		S3Bucket:       cfg.S3Bucket,
		S3Key:          cfg.S3AccessKey,
		S3Secret:       cfg.S3SecretKey,
		S3PublicPrefix: cfg.S3PublicPrefix,
	})

	gitClient := git.NewGiteaClient("https://git.int.sktaurus.ru/api/v1")
	aiClient := ai.NewOpenAIClient(cfg.OpenAIAPIURL, cfg.OpenAIAPIKey, cfg.OpenAIModel)
	notifier := notifications.NewTelegramNotifier(cfg.TelegramBotToken)

	// Инициализируем artifactService до processor, так как processor от него зависит
	artifactService := services.NewArtifactService(queries, s3Service, cfg.S3PublicPrefix)

	processor := workers.NewProcessor(queries, gitClient, aiClient, notifier, artifactService)
	workerPool := workers.NewPool(cfg.WorkerPoolSize, processor)
	workerPool.Start()

	oidcService, err := auth.NewOIDCService(auth.OIDCConfig{
		Issuer:       cfg.OIDCIssuer,
		ClientID:     cfg.OIDCClientID,
		ClientSecret: cfg.OIDCClientSecret,
		RedirectURL:  cfg.OIDCRedirectURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC service: %w", err)
	}

	tokenCache := auth.NewTokenCache(1 * time.Hour)
	authService := auth.NewAuthService(oidcService, nil)

	tokenService := services.NewTokenService(queries)
	projectService := services.NewProjectService(queries)
	branchService := services.NewBranchService(queries)
	buildService := services.NewBuildService(queries)
	eventService := services.NewEventService(queries, workerPool, gitClient)

	return &handlers.Dependencies{
		AuthService:     authService,
		ProjectService:  projectService,
		BranchService:   branchService,
		BuildService:    buildService,
		EventService:    eventService,
		ArtifactService: artifactService,
		TokenService:    tokenService,
		TokenValidator:  tokenService,
		Notifier:        notifier,
		OIDCService:     oidcService,
		TokenCache:      tokenCache,
	}, nil
}
