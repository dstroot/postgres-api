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

	"github.com/dstroot/postgres-api/api"
	"github.com/dstroot/postgres-api/routes"
	"github.com/pkg/errors"
)

func run() error {

	// Initialize app
	api, err := api.Initialize()
	if err != nil {
		return errors.Wrap(err, "initialization error")
	}
	defer api.DB.Close()

	// Initialize our routes
	routes.InitializeRoutes(api)

	// App SIGINT or SIGTERM handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// App server error handling
	errChan := make(chan error, 5)

	log.Printf("%s - %s", api.Cfg.HostName, formattedVersion())
	log.Printf("%s - Starting server on port %v...", api.Cfg.HostName, api.Cfg.Port)

	// Run app server
	go func() {
		errChan <- api.Server.ListenAndServe()
	}()

	// Handle errors/graceful shutdown
	for {
		select {
		case err := <-errChan:
			if err != http.ErrServerClosed {
				return errors.Wrap(err, "http server error")
			}
		case <-sigs:
			fmt.Println("")
			log.Printf("%s - Shutdown signal received, exiting...\n", api.Cfg.HostName)
			// shut down gracefully, but wait no longer than 5 seconds before halting
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			api.Server.Shutdown(ctx)
			if err := api.Server.Shutdown(ctx); err != nil {
				return errors.Wrap(err, "server could not shutdown")
			}
			log.Printf("%s - Server gracefully stopped.\n", api.Cfg.HostName)
			os.Exit(0)
		}
	}
}

func main() {
	err := run()
	if err != nil {
		log.Printf("FATAL: %+v\n", err)
		os.Exit(1)
	}
}
