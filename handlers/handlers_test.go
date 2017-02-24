package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dstroot/postgres-api/app"
	model "github.com/dstroot/postgres-api/models"
)

// https://elithrar.github.io/article/testing-http-handlers-go/
// https://www.thepolyglotdeveloper.com/2017/02/unit-testing-golang-application-includes-http/

// initialize will return an initialized app
func initialize() (a app.App) {
	var err error
	a, err = app.Initialize()
	if err != nil {
		log.Fatalf("Expected clean initialization. Got %s", err.Error())
	}

	a.Router.GET("/products", GetProducts(a.DB))
	a.Router.POST("/product", CreateProduct(a.DB))
	a.Router.GET("/product/:id", GetProduct(a.DB))
	a.Router.PUT("/product/:id", UpdateProduct(a.DB))
	a.Router.DELETE("/product/:id", DeleteProduct(a.DB))

	return a
}

func TestGetProducts(t *testing.T) {

	// This test deletes all records from the products table and sends a GET
	// request to the /products end point. We use the executeRequest function
	// to execute the request. We then use the checkResponseCode function to
	// test that the HTTP response code is what we expect. Finally, we check
	// the body of the response and test that it is the textual representation
	// of an empty array.

	// Initialize the app
	a := initialize()
	defer a.DB.Close()

	// define a product and use it's method to nake sure the table exists
	// and it's clear.
	p := model.Product{}
	p.EnsureTableExists(a.DB)
	p.ClearTable(a.DB)

	// Create a request to pass to our handler. We don't have any
	// parameters for now, so we'll pass 'nil' as the third parameter.
	req, _ := http.NewRequest("GET", "/products", nil)

	// Run the request and get the response
	res := executeRequest(a, req)

	// Check the status code is what we expect.
	if status := res.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := "[]"
	if body := res.Body.String(); body != expected {
		t.Errorf("Expected an empty array. Got %s", body)
	}

	// This test ensures we handle start values less than zero.
	req, _ = http.NewRequest("GET", "/products?count=100&start=-5", nil)

	// Run the request and get the response
	res = executeRequest(a, req)

	checkResponseCode(t, http.StatusOK, res.Code)

	// This tests that accessing product with the database closed
	// returns the relevant error and status code 500.
	a.DB.Close()

	req, _ = http.NewRequest("GET", "/products", nil)

	// Run the request and get the response
	res = executeRequest(a, req)

	checkResponseCode(t, http.StatusInternalServerError, res.Code)

	var m map[string]string
	json.Unmarshal(res.Body.Bytes(), &m)
	if m["error"] != "sql: database is closed" {
		t.Errorf("Expected the 'error' key of the response to be set to 'sql: database is closed'. Got '%s'", m["error"])
	}
}

func TestGetProduct(t *testing.T) {

	// This test tries to access a non-existent product at an endpoint and tests two things:
	// 1) That the status code is 404, indicating that the product was not found, and
	// 2) That the response contains an error with the message "Product not found".

	// Initialize app
	a := initialize()
	defer a.DB.Close()

	p := model.Product{}

	p.ClearTable(a.DB)

	req, _ := http.NewRequest("GET", "/product/11", nil)

	// Run the request and get the response
	res := executeRequest(a, req)

	checkResponseCode(t, http.StatusNotFound, res.Code)

	var m map[string]string
	json.Unmarshal(res.Body.Bytes(), &m)
	if m["error"] != "Product not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Product not found'. Got '%s'", m["error"])
	}

	// This test simply adds a product to the table and tests that accessing
	// the relevant endpoint results in an HTTP response that denotes
	// success with status code 200.

	p.ClearTable(a.DB)
	p.AddTestData(a.DB, 1)

	req, _ = http.NewRequest("GET", "/product/1", nil)

	// Run the request and get the response
	res = executeRequest(a, req)

	checkResponseCode(t, http.StatusOK, res.Code)

	// This tests that accessing product with an invalid id returns
	// the relevant error and bad request with status code 400.

	req, _ = http.NewRequest("GET", "/product/skippy", nil)

	// Run the request and get the response
	res = executeRequest(a, req)

	checkResponseCode(t, http.StatusBadRequest, res.Code)

	json.Unmarshal(res.Body.Bytes(), &m)
	if m["error"] != "Invalid product ID" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Invalid product ID'. Got '%s'", m["error"])
	}

	// This tests that accessing product with the database closed returns
	// the relevant error and status code 500.
	a.DB.Close()

	req, _ = http.NewRequest("GET", "/product/1", nil)

	// Run the request and get the response
	res = executeRequest(a, req)

	checkResponseCode(t, http.StatusInternalServerError, res.Code)

	json.Unmarshal(res.Body.Bytes(), &m)
	if m["error"] != "sql: database is closed" {
		t.Errorf("Expected the 'error' key of the response to be set to 'sql: database is closed'. Got '%s'", m["error"])
	}
}

func TestCreateProduct(t *testing.T) {

	// In this test, we manually add a product to the database and then access
	// the relevant endpoint to fetch that product. We test the following things:
	// 1) That the HTTP response has the status code of 201, indicating that a resource was created, and
	// 2) That the response contained a JSON object with contents identical to that of the payload.

	// Initialize app
	a := initialize()
	defer a.DB.Close()

	p := model.Product{}
	p.EnsureTableExists(a.DB)
	p.ClearTable(a.DB)

	payload := []byte(`{"name":"test product","price":11.22}`)

	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(payload))

	// Run the request and get the response
	res := executeRequest(a, req)

	checkResponseCode(t, http.StatusCreated, res.Code)

	var m map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &m)

	if m["name"] != "test product" {
		t.Errorf("Expected product name to be 'test product'. Got '%v'", m["name"])
	}

	if m["price"] != 11.22 {
		t.Errorf("Expected product price to be '11.22'. Got '%v'", m["price"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
	}

	// In this test, we are sending bad data.  We test the following things:
	// 1) That the HTTP response has the status code of 201, indicating that a resource was created, and
	// 2) That the response contained a JSON object with contents identical to that of the payload.

	// name should be a string, not a number
	payload = []byte(`{"name":100,"price":11.22}`)

	req, _ = http.NewRequest("POST", "/product", bytes.NewBuffer(payload))

	// Run the request and get the response
	res = executeRequest(a, req)

	checkResponseCode(t, http.StatusBadRequest, res.Code)

	json.Unmarshal(res.Body.Bytes(), &m)
	if m["error"] != "Invalid request payload" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Invalid request payload'. Got '%s'", m["error"])
	}

	// This tests that accessing product with the database closed returns
	// the relevant error and status code 500.
	a.DB.Close()

	payload = []byte(`{"name":"test product","price":11.22}`)

	req, _ = http.NewRequest("POST", "/product", bytes.NewBuffer(payload))

	// Run the request and get the response
	res = executeRequest(a, req)

	checkResponseCode(t, http.StatusInternalServerError, res.Code)

	json.Unmarshal(res.Body.Bytes(), &m)
	if m["error"] != "sql: database is closed" {
		t.Errorf("Expected the 'error' key of the response to be set to 'sql: database is closed'. Got '%s'", m["error"])
	}

}

func TestUpdateProduct(t *testing.T) {

	// This test begins by adding a product to the database directly. It then uses
	// the end point to update this record. We test the following things:
	// 1) That the status code is 200, indicating success, and
	// 2) That the response contains the JSON representation of the
	//    product with the updated details.

	// Initialize app
	a := initialize()
	defer a.DB.Close()

	p := model.Product{}
	p.EnsureTableExists(a.DB)
	p.ClearTable(a.DB)
	p.AddTestData(a.DB, 1)

	req, _ := http.NewRequest("GET", "/product/1", nil)

	// Run the request and get the response
	res := executeRequest(a, req)

	var originalProduct map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &originalProduct)

	payload := []byte(`{"name":"test product - updated name","price":11.22}`)

	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(payload))

	// Run the request and get the response
	res = executeRequest(a, req)

	checkResponseCode(t, http.StatusOK, res.Code)

	var m map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &m)

	if m["id"] != originalProduct["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalProduct["id"], m["id"])
	}

	if m["name"] == originalProduct["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalProduct["name"], m["name"], m["name"])
	}

	if m["price"] == originalProduct["price"] {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalProduct["price"], m["price"], m["price"])
	}

	// This tests that updating product with an invalid id returns
	// the relevant error and bad request with status code 400.

	req, _ = http.NewRequest("PUT", "/product/skippy", nil)

	// Run the request and get the response
	res = executeRequest(a, req)

	checkResponseCode(t, http.StatusBadRequest, res.Code)

	json.Unmarshal(res.Body.Bytes(), &m)
	if m["error"] != "Invalid product ID" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Invalid product ID'. Got '%s'", m["error"])
	}

	// In this test, we test sending bad data.  We test the following things:
	// 1) That the HTTP response has the status code of 201, indicating that a resource was created, and
	// 2) That the response contained a JSON object with contents identical to that of the payload.

	// name should be a string, not a number
	payload = []byte(`{"name":100,"price":11.22}`)

	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(payload))

	// Run the request and get the response
	res = executeRequest(a, req)

	checkResponseCode(t, http.StatusBadRequest, res.Code)

	json.Unmarshal(res.Body.Bytes(), &m)
	if m["error"] != "Invalid request payload" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Invalid request payload'. Got '%s'", m["error"])
	}

	// This tests that accessing product with the database closed returns
	// the relevant error and status code 500.
	a.DB.Close()

	payload = []byte(`{"name":"test product","price":11.22}`)

	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(payload))

	// Run the request and get the response
	res = executeRequest(a, req)

	checkResponseCode(t, http.StatusInternalServerError, res.Code)

	json.Unmarshal(res.Body.Bytes(), &m)
	if m["error"] != "sql: database is closed" {
		t.Errorf("Expected the 'error' key of the response to be set to 'sql: database is closed'. Got '%s'", m["error"])
	}

}

func TestDeleteProduct(t *testing.T) {

	// In this test, we first create a product and test that it exists. We then
	// use the endpoint to delete the product. Finally we try to access the
	// product at the appropriate endpoint and test that it doesn't exist.

	// Initialize app
	a := initialize()
	defer a.DB.Close()

	p := model.Product{}
	p.EnsureTableExists(a.DB)
	p.ClearTable(a.DB)
	p.AddTestData(a.DB, 1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	// Run the request and get the response
	res := executeRequest(a, req)
	checkResponseCode(t, http.StatusOK, res.Code)

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	// Run the request and get the response
	res = executeRequest(a, req)
	checkResponseCode(t, http.StatusOK, res.Code)

	req, _ = http.NewRequest("GET", "/product/1", nil)
	// Run the request and get the response
	res = executeRequest(a, req)
	checkResponseCode(t, http.StatusNotFound, res.Code)

	// This tests that deleting product with an invalid id returns
	// the relevant error and bad request with status code 400.

	req, _ = http.NewRequest("DELETE", "/product/skippy", nil)
	// Run the request and get the response
	res = executeRequest(a, req)
	checkResponseCode(t, http.StatusBadRequest, res.Code)

	var m map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &m)
	if m["error"] != "Invalid product ID" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Invalid product ID'. Got '%s'", m["error"])
	}

	// This tests that accessing product with the database closed returns
	// the relevant error and status code 500.
	a.DB.Close()

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	// Run the request and get the response
	res = executeRequest(a, req)
	checkResponseCode(t, http.StatusInternalServerError, res.Code)

	json.Unmarshal(res.Body.Bytes(), &m)
	if m["error"] != "sql: database is closed" {
		t.Errorf("Expected the 'error' key of the response to be set to 'sql: database is closed'. Got '%s'", m["error"])
	}
}

// This function executes the request using the application's router
// and returns the response.
func executeRequest(a app.App, req *http.Request) (res *httptest.ResponseRecorder) {

	// We create a ResponseRecorder (which satisfies
	// http.ResponseWriter) to record the response.
	res = httptest.NewRecorder()

	// Call the routers ServeHTTP method directly and pass
	// in our Request and ResponseRecorder.
	a.Router.ServeHTTP(res, req)

	return res
}

// checkResponseCode validates the correct HTML response code
func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}
