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

// Package routes contains our API routes.
package routes

import (
	"encoding/json"
	"net/http"

	"github.com/dstroot/postgres-api/app"
	"github.com/dstroot/postgres-api/handlers"
	"github.com/julienschmidt/httprouter"
)

// InitializeRoutes intializes our routes
func InitializeRoutes(a app.App) {

	a.Router.GET("/products", handlers.GetProducts(a.DB))
	a.Router.POST("/product", handlers.CreateProduct(a.DB))
	a.Router.GET("/product/:id", handlers.GetProduct(a.DB))
	a.Router.PUT("/product/:id", handlers.UpdateProduct(a.DB))
	a.Router.DELETE("/product/:id", handlers.DeleteProduct(a.DB))

	a.Router.GET("/config", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		s, err := json.MarshalIndent(a.Cfg, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write(s)
	})

	a.Router.GET("/stats", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		s, err := json.MarshalIndent(a.Stats.Data(), "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write(s)
	})

	a.Router.GET("/health", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.Write([]byte("."))
	})
}
