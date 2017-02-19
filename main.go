// postgres-api : a very simple example of a golang API on
// top of a Postgres table
//
// This program is written for go 1.8 and takes advantage of the
// ability to drain connections and do a graceful shutdown.
//
// This API uses:
// * Julien Schmidt's httprouter [package](https://github.com/julienschmidt/httprouter).
//   In contrast to the default mux of Go's net/http package, this router
//   supports variables in the routing pattern and matches against the request
//   method. The router is optimized for high performance and a small memory
//   footprint.
// * Negroni middleware [package](https://github.com/urfave/negroni).
//   Negroni is an idiomatic approach to web middleware in Go. It is tiny,
//   non-intrusive, and encourages use of net/http Handlers.
// * Godotenv [package](https://github.com/joho/godotenv) loads env vars
//   from a .env file. Storing configuration in the environment is one of the
//   tenets of a twelve-factor app. But it is not always practical to set
//   environment variables on development machines or continuous integration
//   servers where multiple projects are run. Godotenv load variables from
//   a .env file into ENV when the environment is bootstrapped.
// * Envdecode [package](https://github.com/joeshaw/envdecode). Envdecode
//   uses struct tags to map environment variables to fields, allowing you
//   you use any names you want for environment variables. In this way you
//   load the environment variables into a config struct once and can then
//   use them throughout your program.
// * Errors [package](https://github.com/pkg/errors).  The errors package
//   allows you to add context to the failure path of your code in a way
//   that does not destroy the original value of the error.
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
	"github.com/pkg/errors"
)

func run() error {

	// Create app
	a := api.App{}

	// Initialize app
	err := a.Initialize()
	if err != nil {
		return errors.Wrap(err, "initialization error")
	}
	defer a.DB.Close()

	// App SIGINT or SIGTERM handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// App server error handling
	errChan := make(chan error, 5)

	log.Printf("%s - %s", a.Cfg.HostName, formattedVersion())
	log.Printf("%s - Starting server on port %v...", a.Cfg.HostName, a.Cfg.Port)

	// Run app server
	go func() {
		errChan <- a.Server.ListenAndServe()
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
			log.Printf("%s - Shutdown signal received, exiting...\n", a.Cfg.HostName)
			// shut down gracefully, but wait no longer than 5 seconds before halting
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			a.Server.Shutdown(ctx)
			if err := a.Server.Shutdown(ctx); err != nil {
				return errors.Wrap(err, "server could not shutdown")
			}
			log.Printf("%s - Server gracefully stopped.\n", a.Cfg.HostName)
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
