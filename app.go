package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/dstroot/postgres-api/models"
	env "github.com/joeshaw/envdecode"
	_ "github.com/joho/godotenv/autoload"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

var (
	cfg Config // global configuration
)

// App struct exposes references to the router, server and database that the application uses.
type App struct {
	Router *httprouter.Router
	DB     *sql.DB
	Server *http.Server
}

// Config contains the configuration from environment variables
type Config struct {
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

// To be useful and testable, App will need methods that initialize
// and run the application.

// Initialize will connect to the database
func (a *App) Initialize() (err error) {

	// Read configuration from env variables
	err = env.Decode(&cfg)
	if err != nil {
		return errors.Wrap(err, "configuration decode failed")
	}

	// configure hostame
	cfg.HostName, _ = os.Hostname()

	// Log configuration for debugging
	if cfg.Debug {
		prettyCfg, _ := json.MarshalIndent(cfg, "", "  ")
		log.Printf("Configuration: \n%v", string(prettyCfg))
	}

	connString := "postgres://" + cfg.SQL.User +
		":" + cfg.SQL.Password +
		"@" + cfg.SQL.Host +
		":" + cfg.SQL.Port +
		"/" + cfg.SQL.Database +
		"?sslmode=disable"

	// Connect to the database
	a.DB, err = sql.Open("postgres", connString)
	if err != nil {
		return errors.Wrap(err, "database connection failed")
	}

	// The first actual connection to the underlying datastore will be
	// established lazily, when it's needed for the first time. If you want
	// to check right away that the database is available and accessible
	// (for example, check that you can establish a network connection and log
	// in), use database.DB.Ping().
	err = a.DB.Ping()
	if err != nil {
		return errors.Wrap(err, "error pinging database")
	}

	if cfg.Debug {
		log.Printf("Connection: %s\n", connString)
	}

	/**
	 * Router
	 */

	a.Router = httprouter.New()
	a.initializeRoutes()

	/**
	 * Negroni Middleware Stack
	 */

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())
	n.UseHandler(a.Router)

	/**
	 * Server
	 */

	a.Server = &http.Server{
		Addr:           ":" + cfg.Port,
		Handler:        n, // pass router
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second, // Go 1.8
		MaxHeaderBytes: 1 << 20,
	}

	return nil
}

// Intialize our routes
func (a *App) initializeRoutes() {
	a.Router.GET("/products", a.getProducts)
	a.Router.POST("/product", a.createProduct)
	a.Router.GET("/product/:id", a.getProduct)
	a.Router.PUT("/product/:id", a.updateProduct)
	a.Router.DELETE("/product/:id", a.deleteProduct)
}

// This handler retrieves the id of the product to be fetched from the requested
// URL, and uses the getProduct method, created in the previous section, to
// fetch the details of that product.
//
// If the product is not found, the handler responds with a status code of 404,
// indicating that the requested resource could not be found. If the product
// is found, the handler responds with the product.
//
// This method uses respondWithError and respondWithJSON functions to
// process errors and normal responses. These functions can be implemented as follows:
func (a *App) getProduct(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	id, err := strconv.Atoi(param.ByName("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	p := model.Product{ID: id}
	if err := p.GetProduct(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Product not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

// This handler uses the count and start parameters from the querystring to
// fetch count number of products, starting at position start in the database.
// By default, start is set to 0 and count is set to 10. If these parameters
// aren't provided, this handler will respond with the first 10 products.
func (a *App) getProducts(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	queryValues := r.URL.Query()
	count, _ := strconv.Atoi(queryValues.Get("count"))
	start, _ := strconv.Atoi(queryValues.Get("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	products, err := model.GetProducts(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, products)
}

// This handler assumes that the request body is a JSON object containing the
// details of the product to be created. It extracts that object into a product
// and uses the createProduct method to create a product with these details.
func (a *App) createProduct(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var p model.Product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := p.CreateProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, p)
}

// This handler extracts the product details from the request body. It also
// extracts the id from the URL and uses the id and the body to update the
// product in the database.
func (a *App) updateProduct(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	id, err := strconv.Atoi(param.ByName("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var p model.Product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}
	defer r.Body.Close()
	p.ID = id

	if err := p.UpdateProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

// This handler extracts the id from the requested URL and uses it to delete
// the corresponding product from the database.
func (a *App) deleteProduct(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
	id, err := strconv.Atoi(param.ByName("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	p := model.Product{ID: id}
	if err := p.DeleteProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
