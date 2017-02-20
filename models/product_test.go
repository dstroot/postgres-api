package model

import (
	"database/sql"
	"encoding/json"
	"log"
	"testing"

	env "github.com/joeshaw/envdecode"
	"github.com/pkg/errors"
)

// TODO this should use the API package.

// App struct exposes references to the router, server, database
// and configuration that the application uses.
type App struct {
	DB  *sql.DB
	Cfg config
}

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

// Initialize will connect to the database
func (a *App) Initialize() (err error) {

	// Read configuration from env variables
	err = env.Decode(&a.Cfg)
	if err != nil {
		return errors.Wrap(err, "configuration decode failed")
	}

	// Log configuration for debugging
	if a.Cfg.Debug {
		prettyCfg, _ := json.MarshalIndent(a.Cfg, "", "  ")
		log.Printf("Configuration: \n%v", string(prettyCfg))
	}

	connString := "postgres://" + a.Cfg.SQL.User +
		":" + a.Cfg.SQL.Password +
		"@" + a.Cfg.SQL.Host +
		":" + a.Cfg.SQL.Port +
		"/" + a.Cfg.SQL.Database +
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

	if a.Cfg.Debug {
		log.Printf("Connection: %s\n", connString)
	}

	return nil
}

var a App

func initialize() {
	a = App{}
	err := a.Initialize()
	if err != nil {
		log.Fatal(err)
	}
}

func TestGet(t *testing.T) {

	initialize()
	defer a.DB.Close()

	p := Product{ID: 1}

	p.EnsureTableExists(a.DB)
	p.ClearTable(a.DB)
	p.AddTestData(a.DB, 1)

	err := p.Get(a.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Errorf("Expected a row to be returned. Got %s", err.Error())
			return
		}

		t.Errorf("Error: %v", err)
	}
}

func TestPut(t *testing.T) {

	initialize()
	defer a.DB.Close()

	p := Product{ID: 1}

	p.EnsureTableExists(a.DB)
	p.ClearTable(a.DB)
	p.AddTestData(a.DB, 1)

	// get product
	err := p.Get(a.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Errorf("Expected a row to be returned. Got %s", err.Error())
			return
		}

		t.Errorf("Error: %v", err)
	}

	// save original data
	original := p

	// update product
	p.Name = "new name"
	err1 := p.Put(a.DB)
	if err != nil {
		if err1 == sql.ErrNoRows {
			t.Errorf("Expected a row to be returned. Got %s", err1.Error())
			return
		}

		t.Errorf("Error: %v", err1)
	}

	// get product again
	err2 := p.Get(a.DB)
	if err2 != nil {
		if err2 == sql.ErrNoRows {
			t.Errorf("Expected a row to be returned. Got %s", err2.Error())
			return
		}

		t.Errorf("Error: %v", err2)
	}

	// check values
	if p.ID != original.ID {
		t.Errorf("Expected the id to remain the same (%v). Got %v", original.ID, p.ID)
	}

	if p.Name != "new name" {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", original.Name, "new name", p.Name)
	}

	if p.Price != original.Price {
		t.Errorf("Expected the price to remain the same (%v). Got %v", original.Price, p.Price)
	}
}

func TestDelete(t *testing.T) {

	initialize()
	defer a.DB.Close()

	p := Product{ID: 1}

	p.EnsureTableExists(a.DB)
	p.ClearTable(a.DB)
	p.AddTestData(a.DB, 1)

	// delete product
	err := p.Delete(a.DB)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	// get product > no rows
	err1 := p.Get(a.DB)
	if err1 != nil {
		if err1 == sql.ErrNoRows {
			return
		}
		t.Errorf("Expected no rows to be returned. Got %s", err1.Error())
	}
}

func TestPost(t *testing.T) {

	initialize()
	defer a.DB.Close()

	new := Product{Name: "hello kitty", Price: 14.99}

	new.EnsureTableExists(a.DB)
	new.ClearTable(a.DB)

	// Post a new product
	err := new.Post(a.DB)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	// get product
	p := Product{ID: 1}
	err1 := p.Get(a.DB)
	if err1 != nil {
		t.Errorf("Error: %v", err1)
	}

	if new.ID != p.ID {
		t.Errorf("Expected the id to remain the same (%v). Got %v", new.ID, p.ID)
	}

	if new.Name != p.Name {
		t.Errorf("Expected the name to remain the same (%v). Got '%v'", new.Name, p.Name)
	}

	if new.Price != p.Price {
		t.Errorf("Expected the price to remain the same (%v). Got %v", new.Price, p.Price)
	}
}

func TestGetMany(t *testing.T) {

	initialize()
	defer a.DB.Close()

	p := Product{}

	p.EnsureTableExists(a.DB)
	p.ClearTable(a.DB)
	p.AddTestData(a.DB, 40)

	// get 8 products
	products, err := GetMany(a.DB, 0, 8)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if len(products) != 8 {
		t.Errorf("Length is wrong")
	}

	a.DB.Close()
	_, err1 := GetMany(a.DB, 0, 8)
	if err1.Error() != "sql: database is closed" {
		t.Errorf("Expected the 'error' response to be 'sql: database is closed'. Got '%s'", err1.Error())
	}
}
