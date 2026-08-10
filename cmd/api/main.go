// Command api starts the Goportunitties HTTP server.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KaueChristian/Goportunitties/internal/config"
	"github.com/KaueChristian/Goportunitties/internal/database"
	"github.com/KaueChristian/Goportunitties/internal/logger"
	"github.com/KaueChristian/Goportunitties/internal/repository"
	"github.com/KaueChristian/Goportunitties/internal/router"
	"github.com/KaueChristian/Goportunitties/internal/service"
	"github.com/KaueChristian/Goportunitties/internal/web"
)

// version is stamped at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("server exited with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// run wires the application and blocks until it is asked to stop. Keeping main
// this thin is what makes every failure path return an error instead of calling
// os.Exit from somewhere deep in the setup.
func run() error {
	settings := config.Load()
	log := logger.New(settings.Env)

	db, err := database.Open(settings.DBPath, !settings.IsProduction())
	if err != nil {
		return err
	}
	defer func() {
		if err := database.Close(db); err != nil {
			log.Error("closing database", slog.String("error", err.Error()))
		}
	}()

	handler := router.New(router.Deps{
		Settings: settings,
		Log:      log,
		DB:       db,
		Service:  service.NewOpeningService(repository.NewOpeningRepository(db)),
		Version:  version,
		SPA:      web.SPA(),
	})

	// Explicit timeouts: the zero-value http.Server that router.Run() builds has
	// none, so a single slow client can hold a connection open indefinitely.
	server := &http.Server{
		Addr:              ":" + settings.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		log.Info("http server listening",
			slog.String("addr", server.Addr),
			slog.String("env", settings.Env),
			slog.String("version", version),
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		log.Info("shutdown signal received", slog.Duration("grace", settings.ShutdownTimeout))
	}

	// Stop accepting connections and give in-flight requests a bounded window to
	// finish before the process goes away.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), settings.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	log.Info("server stopped cleanly")
	return nil
}
