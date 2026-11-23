// Command api is the entry point for the Nexpulse HTTP API.
//
// It loads its configuration, opens a Postgres connection pool, starts the
// HTTP server, and shuts down cleanly when the process receives SIGINT
// (Ctrl+C) or SIGTERM.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/JohnMHoyt24/nexpulse/backend/internal/server"
)

func main() {
	// Load local configuration if present. The error is ignored on purpose so
	// the app still starts when that file doesn't exist.
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ctx is cancelled when the process receives SIGINT or SIGTERM. Everything
	// below waits on it to know when to start shutting down.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// pgxpool.New only parses the config and sets up the pool; it does not
	// open a connection yet, so a bad host or password won't fail here.
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer pool.Close()

	// Ping forces a real connection so we fail fast at startup if the database
	// is unreachable, instead of on the first request.
	pingCtx, cancelPing := context.WithTimeout(ctx, 10*time.Second)
	defer cancelPing()
	if err := pool.Ping(pingCtx); err != nil {
		log.Fatalf("unable to ping database: %v", err)
	}

	log.Println("connected to database")

	// The timeouts stop slow or stalled clients from holding connections open
	// forever (ReadHeaderTimeout in particular guards against Slowloris).
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           server.New(pool),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Run the server in its own goroutine so main can block on the shutdown
	// signal. ErrServerClosed is the expected result of calling Shutdown, so
	// it is not treated as a failure.
	go func() {
		log.Printf("Nexpulse API listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Block until a shutdown signal arrives.
	<-ctx.Done()
	log.Println("shutting down...")

	// Give in-flight requests up to 10 seconds to finish. A fresh context is
	// used because ctx has already been cancelled.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	log.Println("server stopped")
}
