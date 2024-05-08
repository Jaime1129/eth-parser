package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	storage := NewStorage()
	rpcClient := &rpcClient{storage}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()

		// start tracking latest blocks
		rpcClient.trackBlocks(ctx)
	}(ctx)

	srv := &http.Server{
		Addr: ":8080",
	}

	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()

		// Start the server
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("listen: %s\n", err)
		}
	}(ctx)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	cancel()
	wg.Wait()
}
