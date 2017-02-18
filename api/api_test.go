package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	model "github.com/dstroot/postgres-api/models"
)

// We define a global variable a that will represent the api we
// want to test.
var a App

// While calling the Initialize method, we assume that
// the access details for the database will be stored in environment
// variables named TEST_DB_USERNAME, TEST_DB_PASSWORD, and TEST_DB_NAME.
func initialize() {
	a = App{}
	err := a.Initialize()
	if err != nil {
		log.Fatal(err)
	}
}

func TestGetProducts(t *testing.T) {

	// This test deletes all records from the products table and sends a GET request to the /products end point. We use the executeRequest function to execute the request. We then use the checkResponseCode function to test that the HTTP response code is what we expect. Finally, we check the body of the response and test that it is the textual representation of an empty array.
	initialize()
	defer a.DB.Close()

	p := model.Product{}
	p.ClearTable(a.DB)

	req, _ := http.NewRequest("GET", "/products", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}

	// This test ensures we handle start values less than zero.
	req, _ = http.NewRequest("GET", "/products?count=100&start=-5", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	// This tests that accessing product with the database closed returns
	// the relevant error and status code 500.
	a.DB.Close()

	req, _ = http.NewRequest("GET", "/products", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusInternalServerError, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "sql: database is closed" {
		t.Errorf("Expected the 'error' key of the response to be set to 'sql: database is closed'. Got '%s'", m["error"])
	}
}

func TestGetProduct(t *testing.T) {

	// This test tries to access a non-existent product at an endpoint and tests two things:
	// 1) That the status code is 404, indicating that the product was not found, and
	// 2) That the response contains an error with the message "Product not found".

	initialize()
	defer a.DB.Close()

	p := model.Product{}
	p.ClearTable(a.DB)

	req, _ := http.NewRequest("GET", "/product/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Product not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Product not found'. Got '%s'", m["error"])
	}

	// This test simply adds a product to the table and tests that accessing
	// the relevant endpoint results in an HTTP response that denotes
	// success with status code 200.

	p.EnsureTableExists(a.DB)
	p.ClearTable(a.DB)
	p.AddTestData(a.DB, 1)

	req, _ = http.NewRequest("GET", "/product/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	// This tests that accessing product with an invalid id returns
	// the relevant error and bad request with status code 400.

	req, _ = http.NewRequest("GET", "/product/skippy", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Invalid product ID" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Invalid product ID'. Got '%s'", m["error"])
	}

	// This tests that accessing product with the database closed returns
	// the relevant error and status code 500.
	a.DB.Close()

	req, _ = http.NewRequest("GET", "/product/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusInternalServerError, response.Code)

	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "sql: database is closed" {
		t.Errorf("Expected the 'error' key of the response to be set to 'sql: database is closed'. Got '%s'", m["error"])
	}
}

func TestCreateProduct(t *testing.T) {

	// In this test, we manually add a product to the database and then access
	// the relevant endpoint to fetch that product. We test the following things:
	// 1) That the HTTP response has the status code of 201, indicating that a resource was created, and
	// 2) That the response contained a JSON object with contents identical to that of the payload.

	initialize()
	defer a.DB.Close()

	p := model.Product{}
	p.ClearTable(a.DB)

	payload := []byte(`{"name":"test product","price":11.22}`)

	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

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
	response = executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Invalid request payload" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Invalid request payload'. Got '%s'", m["error"])
	}

	// This tests that accessing product with the database closed returns
	// the relevant error and status code 500.
	a.DB.Close()

	payload = []byte(`{"name":"test product","price":11.22}`)

	req, _ = http.NewRequest("POST", "/product", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusInternalServerError, response.Code)

	json.Unmarshal(response.Body.Bytes(), &m)
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

	initialize()
	defer a.DB.Close()

	p := model.Product{}
	p.EnsureTableExists(a.DB)
	p.ClearTable(a.DB)
	p.AddTestData(a.DB, 1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	var originalProduct map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalProduct)

	payload := []byte(`{"name":"test product - updated name","price":11.22}`)

	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

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
	response = executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Invalid product ID" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Invalid product ID'. Got '%s'", m["error"])
	}

	// In this test, we test sending bad data.  We test the following things:
	// 1) That the HTTP response has the status code of 201, indicating that a resource was created, and
	// 2) That the response contained a JSON object with contents identical to that of the payload.

	// name should be a string, not a number
	payload = []byte(`{"name":100,"price":11.22}`)

	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Invalid request payload" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Invalid request payload'. Got '%s'", m["error"])
	}

	// This tests that accessing product with the database closed returns
	// the relevant error and status code 500.
	a.DB.Close()

	payload = []byte(`{"name":"test product","price":11.22}`)

	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusInternalServerError, response.Code)

	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "sql: database is closed" {
		t.Errorf("Expected the 'error' key of the response to be set to 'sql: database is closed'. Got '%s'", m["error"])
	}

}

func TestDeleteProduct(t *testing.T) {

	// In this test, we first create a product and test that it exists. We then
	// use the endpoint to delete the product. Finally we try to access the
	// product at the appropriate endpoint and test that it doesn't exist.

	initialize()
	defer a.DB.Close()

	p := model.Product{}
	p.EnsureTableExists(a.DB)
	p.ClearTable(a.DB)
	p.AddTestData(a.DB, 1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/product/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)

	// This tests that deleting product with an invalid id returns
	// the relevant error and bad request with status code 400.

	req, _ = http.NewRequest("DELETE", "/product/skippy", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Invalid product ID" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Invalid product ID'. Got '%s'", m["error"])
	}

	// This tests that accessing product with the database closed returns
	// the relevant error and status code 500.
	a.DB.Close()

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusInternalServerError, response.Code)

	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "sql: database is closed" {
		t.Errorf("Expected the 'error' key of the response to be set to 'sql: database is closed'. Got '%s'", m["error"])
	}
}

// This function executes the request using the application's router
// and returns the response.
func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

// checkResponseCode validates the correct HTML response code
func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}
