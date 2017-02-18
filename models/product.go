package model

import (
	"database/sql"
	"strconv"
	// Postgres driver
	_ "github.com/lib/pq"
)

// When we define a model we
// - Give it CRUD methods
// - Give it a "GetMany" function
// - Give it helper methods:
//   * EnsureTableExists is a function to create the underlying table if it does not exist
//   * ClearTable is a function to clean out the table
//   * AddTestData is a function to create test data

const createProductsTable = `CREATE TABLE IF NOT EXISTS products
(
id SERIAL,
name TEXT NOT NULL,
price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
CONSTRAINT products_pkey PRIMARY KEY (id)
)`

// Product represents products
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

/**
 * CRUD Methods
 */

// Get gets one product by id
func (p *Product) Get(db *sql.DB) error {
	err := db.QueryRow("SELECT name, price FROM products WHERE id=$1",
		p.ID).Scan(&p.Name, &p.Price)

	return err
}

// Put updates one product by id
func (p *Product) Put(db *sql.DB) error {
	_, err := db.Exec("UPDATE products SET name=$1, price=$2 WHERE id=$3", p.Name, p.Price, p.ID)

	return err
}

// Delete deletes one product by id
func (p *Product) Delete(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM products WHERE id=$1", p.ID)

	return err
}

// Post creates a new product
func (p *Product) Post(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO products(name, price) VALUES($1, $2) RETURNING id",
		p.Name, p.Price).Scan(&p.ID)

	return err
}

// GetMany fetches a list of products
func GetMany(db *sql.DB, start, count int) ([]Product, error) {
	rows, err := db.Query(
		"SELECT id, name, price FROM products LIMIT $1 OFFSET $2",
		count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	products := []Product{}

	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

/**
 * Helpers
 */

// EnsureTableExists makes sure that the table we need for testing is available.
func (p *Product) EnsureTableExists(db *sql.DB) error {
	if _, err := db.Exec(createProductsTable); err != nil {
		return err
	}

	return nil
}

// ClearTable to clean the table up.
func (p *Product) ClearTable(db *sql.DB) error {
	if _, err := db.Exec("DELETE FROM products"); err != nil {
		return err
	}
	if _, err := db.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1"); err != nil {
		return err
	}
	return nil
}

// AddTestData is used to add one or more records into the table for testing.
func (p *Product) AddTestData(db *sql.DB, count int) error {
	if count < 1 {
		count = 1
	}

	var err error
	for i := 0; i < count; i++ {
		_, err = db.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Product "+strconv.Itoa(i), (i+1.0)*10)
		if err != nil {
			return err
		}
	}

	return nil
}
