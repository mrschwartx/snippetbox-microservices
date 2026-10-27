package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"gorm.io/driver/mysql"

	"snippetbox/backend-user/internal/app"
	"snippetbox/backend-user/internal/config"
	"snippetbox/backend-user/internal/database"
	"snippetbox/backend-user/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("backend stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Connect(ctx, mysql.Open(cfg.DBDSN), database.Retry{
		Attempts: cfg.DBConnectRetries,
		Delay:    cfg.DBConnectRetryDelay,
	}, logger)
	if err != nil {
		return err
	}
	defer func() {
		if err := database.Close(db); err != nil {
			logger.Error("closing database", "error", err)
		}
	}()

	if err := database.Migrate(db); err != nil {
		return err
	}
	logger.Info("database migrated")

	listener, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return err
	}

	srv := server.New(app.NewHandler(app.Dependencies{Config: cfg, DB: db, Logger: logger}))
	logger.Info("backend listening", "addr", listener.Addr().String(), "env", cfg.Env)

	return server.Serve(ctx, srv, listener, logger)
}
