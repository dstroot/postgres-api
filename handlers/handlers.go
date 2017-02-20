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

// Package handler contains our route handlers.  It uses the models
// package.
package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dstroot/postgres-api/models"
	// Load environment vars
	_ "github.com/joho/godotenv/autoload"
	"github.com/julienschmidt/httprouter"
	// Postgres driver
	_ "github.com/lib/pq"
)

// GetProduct retrieves the id of the product to be fetched from the requested
// URL, and uses the getProduct method, created in the previous section, to
// fetch the details of that product.
//
// If the product is not found, the handler responds with a status code of 404,
// indicating that the requested resource could not be found. If the product
// is found, the handler responds with the product.
//
// This method uses respondWithError and respondWithJSON functions to
// process errors and normal responses. These functions can be implemented as follows:
func GetProduct(db *sql.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
		id, err := strconv.Atoi(param.ByName("id"))
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid product ID")
			return
		}

		p := model.Product{ID: id}
		if err := p.Get(db); err != nil {
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
}

// GetProducts uses the count and start parameters from the querystring to
// fetch count number of products, starting at position start in the database.
// By default, start is set to 0 and count is set to 10. If these parameters
// aren't provided, this handler will respond with the first 10 products.
func GetProducts(db *sql.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		queryValues := r.URL.Query()
		count, _ := strconv.Atoi(queryValues.Get("count"))
		start, _ := strconv.Atoi(queryValues.Get("start"))

		if count > 50 || count < 1 {
			count = 50
		}
		if start < 0 {
			start = 0
		}

		products, err := model.GetMany(db, start, count)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, products)
	}
}

// CreateProduct assumes that the request body is a JSON object containing the
// details of the product to be created. It extracts that object into a product
// and uses the createProduct method to create a product with these details.
func CreateProduct(db *sql.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var p model.Product
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&p); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		defer r.Body.Close()

		if err := p.Post(db); err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithJSON(w, http.StatusCreated, p)
	}
}

// UpdateProduct extracts the product details from the request body. It also
// extracts the id from the URL and uses the id and the body to update the
// product in the database.
func UpdateProduct(db *sql.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
		id, err := strconv.Atoi(param.ByName("id"))
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid product ID")
			return
		}

		var p model.Product
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&p); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		defer r.Body.Close()
		p.ID = id

		if err := p.Put(db); err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, p)
	}
}

// DeleteProduct extracts the id from the requested URL and uses it to delete
// the corresponding product from the database.
func DeleteProduct(db *sql.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, param httprouter.Params) {
		id, err := strconv.Atoi(param.ByName("id"))
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid product ID")
			return
		}

		p := model.Product{ID: id}
		if err := p.Delete(db); err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
	}
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
