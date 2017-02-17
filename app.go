package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	env "github.com/joeshaw/envdecode"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

var (
	cfg Config // global configuration
)

// App struct exposes references to the router and the database that the application uses.
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// Config contains the configuration from environment variables
type Config struct {
	Debug bool   `env:"DEBUG,default=true"`
	Port  string `env:"PORT,default=8000"`

	SQL struct {
		// Host     string `env:"MSSQL_HOST,default=localhost"`
		// Port     string `env:"MSSQL_PORT,default=1433"`
		User     string `env:"SQL_USER,default=postgres"`
		Password string `env:"SQL_PASSWORD,default=mysecretpassword"`
		Database string `env:"SQL_DATABASE,default=products"`
	}
}

// To be useful and testable, App will needgo methods that initialize and run the application.

// Initialize will connect to the database
func (a *App) Initialize() (err error) {

	// Read configuration from env variables
	err = env.Decode(&cfg)
	if err != nil {
		return errors.Wrap(err, "configuration decode failed")
	}

	// log configuration for debugging
	if cfg.Debug {
		prettyCfg, _ := json.MarshalIndent(cfg, "", "  ")
		log.Printf("Configuration: \n%v", string(prettyCfg))
	}

	a.DB, err = sql.Open("postgres", "postgres://"+cfg.SQL.User+":"+cfg.SQL.Password+"@localhost/"+cfg.SQL.Database+"?sslmode=disable")
	if err != nil {
		return errors.Wrap(err, "database connection failed")
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()

	return nil
}

// Run will run the app
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(":"+addr, a.Router))
}

// Intialize our routes
func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/products", a.getProducts).Methods("GET")
	a.Router.HandleFunc("/product", a.createProduct).Methods("POST")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.getProduct).Methods("GET")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.updateProduct).Methods("PUT")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.deleteProduct).Methods("DELETE")
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
func (a *App) getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	p := product{ID: id}
	if err := p.getProduct(a.DB); err != nil {
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
func (a *App) getProducts(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	products, err := getProducts(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, products)
}

// This handler assumes that the request body is a JSON object containing the
// details of the product to be created. It extracts that object into a product
// and uses the createProduct method to create a product with these details.
func (a *App) createProduct(w http.ResponseWriter, r *http.Request) {
	var p product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := p.createProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, p)
}

// This handler extracts the product details from the request body. It also
// extracts the id from the URL and uses the id and the body to update the
// product in the database.
func (a *App) updateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var p product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}
	defer r.Body.Close()
	p.ID = id

	if err := p.updateProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

// This handler extracts the id from the requested URL and uses it to delete
// the corresponding product from the database.
func (a *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Product ID")
		return
	}

	p := product{ID: id}
	if err := p.deleteProduct(a.DB); err != nil {
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
