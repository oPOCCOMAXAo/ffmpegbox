package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/opoccomaxao/ffmpegbox/internal/config"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger := initLogger(cfg)
	slog.SetDefault(logger)

	logger.Info("Configuration loaded successfully",
		"config_path", *configPath,
		"bind_address", cfg.Server.BindAddress,
		"port", cfg.Server.Port,
		"auth_enabled", cfg.Auth.Enabled,
		"client_count", len(cfg.Auth.Clients),
		"ffmpeg_binary", cfg.FFmpeg.BinaryPath,
		"temp_dir", cfg.Storage.TempDir,
		"database_path", cfg.Storage.DatabasePath,
	)

	logger.Info("Server initialization complete. Ready to start.")
	os.Exit(0)
}

func initLogger(cfg *config.Config) *slog.Logger {
	var level slog.Level
	switch cfg.Logging.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if cfg.Logging.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
