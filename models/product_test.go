package model

import (
	"database/sql"
	"testing"

	"github.com/dstroot/postgres-api/api"
)

func TestGet(t *testing.T) {

	// Initialize app
	a, err := api.Initialize()
	if err != nil {
		t.Errorf("Expected clean initialization. Got %s", err.Error())
	}
	defer a.DB.Close()

	p := Product{ID: 1}

	p.EnsureTableExists(a.DB)
	p.ClearTable(a.DB)
	p.AddTestData(a.DB, 1)

	err1 := p.Get(a.DB)
	if err1 != nil {
		if err1 == sql.ErrNoRows {
			t.Errorf("Expected a row to be returned. Got %s", err1.Error())
			return
		}

		t.Errorf("Error: %v", err1)
	}

	p.ClearTable(a.DB)
}

func TestPut(t *testing.T) {

	// Initialize app
	a, err := api.Initialize()
	if err != nil {
		t.Errorf("Expected clean initialization. Got %s", err.Error())
	}
	defer a.DB.Close()

	p := Product{ID: 1}

	p.EnsureTableExists(a.DB)
	p.ClearTable(a.DB)
	p.AddTestData(a.DB, 1)

	// get product
	err1 := p.Get(a.DB)
	if err1 != nil {
		if err1 == sql.ErrNoRows {
			t.Errorf("Expected a row to be returned. Got %s", err1.Error())
			return
		}

		t.Errorf("Error: %v", err1)
	}

	// save original data
	original := p

	// update product
	p.Name = "new name"
	err2 := p.Put(a.DB)
	if err2 != nil {
		if err2 == sql.ErrNoRows {
			t.Errorf("Expected a row to be returned. Got %s", err2.Error())
			return
		}

		t.Errorf("Error: %v", err2)
	}

	// get product again
	err3 := p.Get(a.DB)
	if err3 != nil {
		if err3 == sql.ErrNoRows {
			t.Errorf("Expected a row to be returned. Got %s", err3.Error())
			return
		}

		t.Errorf("Error: %v", err3)
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

	p.ClearTable(a.DB)
}

func TestDelete(t *testing.T) {

	// Initialize app
	a, err := api.Initialize()
	if err != nil {
		t.Errorf("Expected clean initialization. Got %s", err.Error())
	}
	defer a.DB.Close()

	p := Product{ID: 1}

	p.EnsureTableExists(a.DB)
	p.ClearTable(a.DB)
	p.AddTestData(a.DB, 1)

	// delete product
	err1 := p.Delete(a.DB)
	if err1 != nil {
		t.Errorf("Error: %v", err1)
	}

	// get product > no rows
	err2 := p.Get(a.DB)
	if err2 != nil {
		if err2 == sql.ErrNoRows {
			return
		}
		t.Errorf("Expected no rows to be returned. Got %s", err2.Error())
	}

	p.ClearTable(a.DB)
}

func TestPost(t *testing.T) {

	// Initialize app
	a, err := api.Initialize()
	if err != nil {
		t.Errorf("Expected clean initialization. Got %s", err.Error())
	}
	defer a.DB.Close()

	new := Product{Name: "hello kitty", Price: 14.99}

	new.EnsureTableExists(a.DB)
	new.ClearTable(a.DB)

	// Post a new product
	err1 := new.Post(a.DB)
	if err1 != nil {
		t.Errorf("Error: %v", err1)
	}

	// get product
	p := Product{ID: 1}
	err2 := p.Get(a.DB)
	if err2 != nil {
		t.Errorf("Error: %v", err2)
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

	p.ClearTable(a.DB)
}

func TestGetMany(t *testing.T) {

	// Initialize app
	a, err := api.Initialize()
	if err != nil {
		t.Errorf("Expected clean initialization. Got %s", err.Error())
	}
	defer a.DB.Close()

	p := Product{}

	p.EnsureTableExists(a.DB)
	p.ClearTable(a.DB)
	p.AddTestData(a.DB, 40)

	// get 8 products
	products, err1 := GetMany(a.DB, 0, 8)
	if err1 != nil {
		t.Errorf("Error: %v", err1)
	}

	if len(products) != 8 {
		t.Errorf("Length is wrong")
	}

	p.ClearTable(a.DB)

	a.DB.Close()
	_, err2 := GetMany(a.DB, 0, 8)
	if err2.Error() != "sql: database is closed" {
		t.Errorf("Expected the 'error' response to be 'sql: database is closed'. Got '%s'", err2.Error())
	}
}
