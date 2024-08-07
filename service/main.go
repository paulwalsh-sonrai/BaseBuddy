package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	workerCount = 5
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Start HTTP server
	go startHTTPServer(ctx)

	// Start worker threads
	var wg sync.WaitGroup
	startWorkerThreads(ctx, &wg, workerCount)

	// Wait for termination signal
	<-signalChan
	log.Println("Received termination signal. Shutting down...")

	// Cancel context to stop worker threads
	cancel()

	// Wait for all workers to finish
	wg.Wait()
	log.Println("All workers stopped.")
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world!")
}

func startHTTPServer(ctx context.Context) {
	server := &http.Server{
		Addr:    ":8080",
		Handler: http.HandlerFunc(handler),
	}

	// Run server in a separate goroutine
	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Shutdown the server
	log.Println("Shutting down HTTP server...")
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("HTTP server Shutdown Failed:%+v", err)
	}
	log.Println("HTTP server stopped.")
}

func startWorkerThreads(ctx context.Context, wg *sync.WaitGroup, count int) {
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			log.Printf("Worker %d started", id)
			for {
				select {
				case <-ctx.Done():
					log.Printf("Worker %d stopping", id)
					return
				default:
					// Simulate work
					time.Sleep(1 * time.Second)
					log.Printf("Worker %d doing work", id)
				}
			}
		}(i)
	}
}
