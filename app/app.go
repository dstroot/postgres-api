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

// Package app contains our application and initialization function.
package app

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/thirdparty/tollbooth_negroni"
	"github.com/dstroot/postgres-api/middleware/connlimit"
	env "github.com/joeshaw/envdecode"
	"github.com/thoas/stats"
	"github.com/urfave/negroni"
	// Load environment vars
	_ "github.com/joho/godotenv/autoload"
	"github.com/julienschmidt/httprouter"
	// Postgres driver
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

// App struct holds the router, server, database
// and configuration that the application uses.
type App struct {
	Router *httprouter.Router
	DB     *sql.DB
	Server *http.Server
	Stats  *stats.Stats
	Cfg    config
}

// config holds the system configuration
type config struct {
	HostName string
	Debug    bool   `env:"DEBUG,default=true"`
	Port     string `env:"PORT,default=8000"`

	SQL struct {
		Host     string `env:"SQL_HOST,default=localhost"`
		Port     string `env:"SQL_PORT,default=5432"`
		User     string `env:"SQL_USER,default=postgres"`
		Password string `env:"SQL_PASSWORD,default=mysecretpassword"`
		Database string `env:"SQL_DATABASE,default=products"`
	}
}

// Initialize will populate the configuration, connect to the database,
// and instantiate the router and server and then return an app.
func Initialize() (app App, err error) {

	/**
	 * Configuration
	 */

	// Read configuration from env variables
	err = env.Decode(&app.Cfg)
	if err != nil {
		return app, errors.Wrap(err, "configuration decode failed")
	}

	// configure hostame
	app.Cfg.HostName, _ = os.Hostname()

	/**
	 * Database
	 */

	connString := "postgres://" + app.Cfg.SQL.User +
		":" + app.Cfg.SQL.Password +
		"@" + app.Cfg.SQL.Host +
		":" + app.Cfg.SQL.Port +
		"/" + app.Cfg.SQL.Database +
		"?sslmode=disable"

	// Connect to the database
	app.DB, err = sql.Open("postgres", connString)
	if err != nil {
		return app, errors.Wrap(err, "database connection failed")
	}

	// The first actual connection to the underlying datastore will be
	// established lazily, when it's needed for the first time. If you want
	// to check right away that the database is available and accessible
	// (for example, check that you can establish a network connection and log
	// in), use database.DB.Ping().
	err = app.DB.Ping()
	if err != nil {
		return app, errors.Wrap(err, "error pinging database")
	}

	/**
	 * Router
	 */

	app.Router = httprouter.New()

	/**
	 * Negroni Middleware Stack
	 */

	// Standard stack, recovery and logging
	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())

	// setup stats https://github.com/thoas/stats
	app.Stats = stats.New()
	n.Use(app.Stats)

	// Connections limiter
	// Manage connections before rate?
	n.Use(connlimit.MaxAllowed(50))

	// Rate limiter
	limiter := tollbooth.NewLimiter(50, time.Second)
	n.Use(tollbooth_negroni.LimitHandler(limiter))

	n.UseHandler(app.Router)

	/**
	 * Server
	 */

	app.Server = &http.Server{
		Addr:           ":" + app.Cfg.Port,
		Handler:        n, // pass in router
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return app, nil
}
