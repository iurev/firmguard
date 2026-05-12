package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"firmguard/internal/database"
	"firmguard/internal/repository"
	"firmguard/internal/server"
	"firmguard/internal/worker"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

func gracefulShutdown(apiServer *http.Server, riverClient *river.Client[pgx.Tx], db database.Service, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop() // Allow Ctrl+C to force shutdown

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if riverClient != nil {
		if err := riverClient.Stop(ctx); err != nil {
			log.Printf("River client stop error: %v", err)
		}
	}

	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	if err := db.Close(); err != nil {
		log.Printf("DB close error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {
	ctx := context.Background()

	// Initialize Database
	db := database.New()
	dbPool := db.GetPool()

	// Initialize River
	workers := river.NewWorkers()
	scanRepo := repository.NewFirmwareScanRepository(db.GetPool())
	vulnRepo := repository.NewVulnerabilityRepository(db.GetPool())
	river.AddWorker(workers, worker.NewFirmwareAnalysisWorker(scanRepo, vulnRepo))

	riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 10},
		},
		Workers:      workers,
		ErrorHandler: &worker.SentryMock{},
	})
	if err != nil {
		log.Fatalf("failed to create river client: %v", err)
	}

	if err := riverClient.Start(ctx); err != nil {
		log.Fatalf("failed to start river client: %v", err)
	}

	server := server.NewServer(riverClient)

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, riverClient, db, done)

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
