package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	a := App{}

	err := a.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer a.DB.Close()

	// SIGINT or SIGTERM handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Error handling
	errChan := make(chan error, 10)

	log.Printf("%s - Starting server on port %v...", cfg.HostName, cfg.Port)

	// Run server
	go func() {
		errChan <- a.Server.ListenAndServe()
	}()

	// Handle channels/graceful shutdown
	for {
		select {
		case err := <-errChan:
			if err != http.ErrServerClosed {
				log.Fatalf("listen: %s\n", err)
			}
		case <-sigs:
			fmt.Println("")
			log.Printf("%s - Shutdown signal received, exiting...\n", cfg.HostName)
			// shut down gracefully, but wait no longer than 5 seconds before halting
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			a.Server.Shutdown(ctx)
			if err := a.Server.Shutdown(ctx); err != nil {
				log.Fatalf("Server could not shutdown: %v", err)
			}
			log.Printf("%s - Server gracefully stopped.\n", cfg.HostName)
			os.Exit(0)
		}
	}
}
