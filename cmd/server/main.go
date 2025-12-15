package main

import (
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gopkg.in/natefinch/lumberjack.v2"

	"vpnbot-core-go/internal/api"
	"vpnbot-core-go/internal/config"
	"vpnbot-core-go/internal/nodeclient"
	"vpnbot-core-go/internal/repository"
	"vpnbot-core-go/internal/service"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	var logLevel slog.Level
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Setup log output (stdout or file with rotation)
	var logWriter io.Writer
	if cfg.LogOutput == "file" {
		// Ensure log directory exists
		logDir := filepath.Dir(cfg.LogFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Fatalf("Failed to create log directory: %v", err)
		}

		// Setup log rotation
		logWriter = &lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    cfg.LogMaxSize,    // megabytes
			MaxBackups: cfg.LogMaxBackups, // number of backups
			MaxAge:     cfg.LogMaxAge,     // days
			Compress:   cfg.LogCompress,   // compress old files
		}
		log.Printf("Logging to file: %s (max size: %dMB, backups: %d, max age: %d days, compress: %v)",
			cfg.LogFile, cfg.LogMaxSize, cfg.LogMaxBackups, cfg.LogMaxAge, cfg.LogCompress)
	} else {
		// Default: stdout (good for Docker)
		logWriter = os.Stdout
	}

	logger := slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	// Initialize repository
	repo, err := repository.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	// Initialize node client
	nodeClient := nodeclient.New(cfg.XrayNodeToken, cfg.NodeTimeout)

	// Initialize service
	svc := service.New(repo, nodeClient, logger, cfg.RequestTimeout)

	// Initialize handlers
	handlers := api.NewHandlers(repo, svc, nodeClient, logger)

	// Setup Fiber app
	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
		AppName:               "VPN Bot Core",
		ServerHeader:          "VPN Bot Core",
		ErrorHandler:          customErrorHandler,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept, X-Api-Key",
	}))

	// Setup routes
	api.SetupRoutes(app, handlers, cfg.APIKey)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("Shutting down server...")
		if err := app.Shutdown(); err != nil {
			logger.Error("Server forced to shutdown", slog.String("error", err.Error()))
		}
	}()

	// Start server
	logger.Info("Server starting", slog.String("port", cfg.Port))
	if err := app.Listen(":" + cfg.Port); err != nil {
		logger.Error("Server failed to start", slog.String("error", err.Error()))
	}

	logger.Info("Server stopped")
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}

