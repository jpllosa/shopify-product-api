package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"

	"shopify-product-api/config"
	"shopify-product-api/router"
)

func main() {
	config, err := config.Load("./conf/config.json")
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", router.GetStatus)

	r.Mount("/products", router.GetProducts(config))

	http.ListenAndServe(fmt.Sprintf(":%v", config.Port), r)
}
