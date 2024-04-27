package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	router := chi.NewRouter()
	if err := petStoreApi.SetupRoutes(router); err != nil {
		panic(err)
	}
	_ = http.ListenAndServe(":8080", router)
	/*
		Default redoc UI will be served on http://localhost:8080/docs
		Swagger style UI will be served on http://localhost:8080/swagger
		Rapidoc style UI will be served on http://localhost:8080/rapidoc
	*/
}
