package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	router := chi.NewRouter()
	if err := myApi.SetupRoutes(router); err != nil {
		panic(err)
	}
	_ = http.ListenAndServe(":8080", router)
}
