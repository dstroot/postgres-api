// The MIT License (MIT)
//
// Copyright (c) 2017 Daniel J. Stroot
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package main initializes the application and runs it using go1.8
// features to stop it gracefully and allow connections to drain.
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

	"github.com/dstroot/postgres-api/app"
	"github.com/dstroot/postgres-api/routes"
	"github.com/pkg/errors"
)

func run() error {

	// Initialize app
	api, err := app.Initialize()
	if err != nil {
		return errors.Wrap(err, "initialization error")
	}
	defer api.DB.Close()

	// Initialize our routes
	routes.InitializeRoutes(api)

	// App SIGINT or SIGTERM handling
	// use a buffered channel or risk missing the signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// App server error handling
	errChan := make(chan error, 5)

	log.Printf("%s - %s", api.Cfg.HostName, formattedVersion())
	log.Printf("%s - Starting server on port %v...", api.Cfg.HostName, api.Cfg.Port)

	// Run API server
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
