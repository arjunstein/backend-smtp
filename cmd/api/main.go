package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/arjunstein/backend-smtp/docs"
	"github.com/arjunstein/backend-smtp/internal/config"
	"github.com/arjunstein/backend-smtp/internal/handler"
	"github.com/arjunstein/backend-smtp/internal/middleware"
	"github.com/arjunstein/backend-smtp/internal/model"
	"github.com/arjunstein/backend-smtp/internal/repository"
	"github.com/arjunstein/backend-smtp/internal/service"
	"github.com/arjunstein/backend-smtp/internal/service/mailer"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// @title Centralized SMTP Email Sender Service
// @version 1.0
// @description Production-grade centralized SMTP agent and email sending engine.
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @tag.name Health
// @tag.description Service health and readiness operations

// @tag.name SMTP Config
// @tag.description Management and testing of SMTP server configurations

// @tag.name Emails
// @tag.description Centralized email dispatch endpoint

func main() {
	// Configure structured JSON logging to stdout for cloud-native observability
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(jsonHandler))

	// 1. Load configuration
	cfg := config.Load()

	// 2. Initialize Database connection
	log.Printf("[Database] Connecting to MySQL at %s:%s (database: %s)...", cfg.DBHost, cfg.DBPort, cfg.DBName)
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		log.Fatalf("[Database] Failed to connect to MySQL: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("[Database] Failed to get generic database handle: %v", err)
	}
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	// Run auto migrations for smtp_configs
	log.Println("[Database] Executing database auto-migrations...")
	if err := db.AutoMigrate(&model.SMTPConfig{}); err != nil {
		log.Fatalf("[Database] Migration failure: %v", err)
	}
	log.Println("[Database] Auto-migrations applied successfully.")

	// 3. Initialize Clean Architecture components
	mailerEngine := mailer.NewMailer()
	smtpRepo := repository.NewSMTPConfigRepository(db)

	smtpService := service.NewSMTPConfigService(smtpRepo, mailerEngine)
	emailService := service.NewEmailService(smtpRepo, mailerEngine)

	healthHandler := handler.NewHealthHandler()
	smtpConfigHandler := handler.NewSMTPConfigHandler(smtpService)
	emailHandler := handler.NewEmailHandler(emailService)

	// 4. Initialize Chi Router & Middleware
	r := chi.NewRouter()
	r.Use(middleware.RequestLogger)
	r.Use(middleware.CORS())
	r.Use(chimiddleware.Recoverer)

	// Public Health Check
	r.Get("/health", healthHandler.CheckHealth)

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	r.Get("/documentation/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	r.Get("/documentation", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/documentation/index.html", http.StatusMovedPermanently)
	})

	// Protected API Routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.APIKeyAuth(cfg.APIKey))

		// SMTP Config endpoints
		r.Route("/smtp-configs", func(r chi.Router) {
			r.Post("/", smtpConfigHandler.Create)
			r.Get("/", smtpConfigHandler.List)
			r.Get("/{id}", smtpConfigHandler.GetByID)
			r.Put("/{id}", smtpConfigHandler.Update)
			r.Delete("/{id}", smtpConfigHandler.Delete)
			r.Put("/{id}/default", smtpConfigHandler.SetDefault)
			r.Post("/{id}/test", smtpConfigHandler.TestConnection)
		})

		// Centralized Email sending endpoint
		r.Route("/emails", func(r chi.Router) {
			r.Post("/send", emailHandler.Send)
		})
	})

	// 5. Start HTTP Server with Graceful Shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[HTTP] Centralized SMTP Sender Service listening on :%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[HTTP] Server listen failed: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[Lifecycle] Shutdown signal received, terminating server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[Lifecycle] Server forced shutdown: %v", err)
	}

	_ = sqlDB.Close()
	log.Println("[Lifecycle] Server stopped cleanly.")
}
